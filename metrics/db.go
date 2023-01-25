package metrics

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

const (
	DB_METRIC_GROUP_TYPE_ENUM_NAME = "metric_group_type_enum"
	DB_METRIC_TABLE_NAME           = "short_metrics"
)

var (
	MDBctx                *metricDBContext
	dbMetricGroupTypeName = DB_METRIC_GROUP_TYPE_ENUM_NAME
	dbMetricTableName     = DB_METRIC_TABLE_NAME
)

type metricDBContext struct {
	ctx           context.Context
	conn          *pgx.Conn
	tableName     string
	groupTypeName string
	ready         bool
}

func init() {
	MDBctx = &metricDBContext{
		ctx:           context.Background(),
		tableName:     dbMetricTableName,
		groupTypeName: dbMetricGroupTypeName,
	}
	var err error
	//
	if metricDBEnv, ok := os.LookupEnv("SH0R7_METRIC_DB_PATH"); ok {
		MDBctx.conn, err = pgx.Connect(MDBctx.ctx, metricDBEnv)
		if err != nil {
			panic(errors.Wrapf(err, "failed to connect to db <%s>", metricDBEnv))
		}
		if envDevTableNamePrefix := os.Getenv("SH0R7_METRIC_DB_TABLE_DEV_PREFIX"); envDevTableNamePrefix != "" {
			MDBctx.tableName = envDevTableNamePrefix + dbMetricTableName
		}
		if envDevGroupTypePrefix := os.Getenv("SH0R7_METRIC_DB_GROUP_TYPE_DEV_PREFIX"); envDevGroupTypePrefix != "" {
			MDBctx.groupTypeName = envDevGroupTypePrefix + dbMetricGroupTypeName
		}
		if devEnvSuffix := os.Getenv("__DEV_ENV"); devEnvSuffix != "" {
			MDBctx.tableName += devEnvSuffix
			MDBctx.groupTypeName += devEnvSuffix
		}

		err = MDBctx.checkCreateMetricTable()
		if err != nil {
			panic(errors.Wrapf(err, "failed to check/create metric table in db <%s>", MDBctx.tableName))
		}
		MDBctx.ready = true
	}
}

func (m *metricDBContext) checkCreateMetricTable() error {
	check, err := MDBctx.isShortMetricsTableExists()
	if err != nil {
		return errors.Wrapf(err, "failed checking if metric table exists <%s>", MDBctx.tableName)
	}
	if check {
		return nil
	}
	err = m.checkCreateMetricGroupType()
	if err != nil {
		return errors.Wrapf(err, "failed check/create metric group type exists <%s>", MDBctx.groupTypeName)
	}
	return m.createMetricTable()
}

func (m *metricDBContext) checkCreateMetricGroupType() error {
	check, err := MDBctx.isShortMetricsGroupTypeExists()
	if err != nil {
		return errors.Wrapf(err, "failed checking if metric group type exists <%s>", MDBctx.tableName)
	}
	if check {
		return nil
	}

	return m.createMetricGroupType()
}

func (m *metricDBContext) isShortMetricsGroupTypeExists() (bool, error) {
	/*
	   select
	  	   n.nspname as enum_schema,
	       t.typname as enum_name,
	       e.enumlabel as enum_value
	   from pg_type t
	      join pg_catalog.pg_enum e on t.oid = e.enumtypid
	      join pg_catalog.pg_namespace n ON n.oid = t.typnamespace
	   where
	   	t.typname = 'metricgroup'
	*/
	field := "enum_value"
	query := fmt.Sprintf(`select
	       e.enumlabel as %s
	   from pg_type t
	      join pg_catalog.pg_enum e on t.oid = e.enumtypid
	      join pg_catalog.pg_namespace n ON n.oid = t.typnamespace
	   where
	   	t.typname = '%s'`, field, m.groupTypeName)
	log.Printf("query to submit:\n[%s]\n", query)
	rows, err := m.conn.Query(m.ctx, query)
	if err != nil {
		if err != pgx.ErrNoRows {
			return false, errors.Wrapf(err, "failed query <%s>", query)
		}
		return false, nil
	}
	enumsLocal := ListMetricGroupsString()
	enumsDB := []string{}
	for rows.Next() {

		if err == pgx.ErrNoRows { // TODO:  not sure on this check
			return false, nil
		}
		v, err := rows.Values()
		if err != nil {
			if err == pgx.ErrNoRows { // TODO:  not sure on this check
				return false, nil
			}
			return false, err
		}
		name, ok := v[0].(string)
		if !ok {
			panic(errors.Errorf("something weird - cant convert to string: %+#v", v[0]))
		}
		enumsDB = append(enumsDB, name)
	}
	if len(enumsDB) == 0 {
		return false, nil
	}
	if len(enumsDB) != len(enumsLocal) {
		err = errors.Errorf("mismatch on group types count: db <%+#v> vs program <%+#v>", enumsDB, enumsLocal)
		log.Println(err.Error())
		return false, err
	}
	sort.Strings(enumsDB)
	sort.Strings(enumsLocal)
	// log.Printf("enumDB:\n%s\n", strings.Join(enumsDB, "#"))
	// log.Printf("enumLocal:\n%s\n", strings.Join(enumsLocal, "#"))
	if strings.Compare(strings.Join(enumsDB, "#"), strings.Join(enumsLocal, "#")) != 0 {
		err = errors.Errorf("mismatch on group types values: db <%+#v> vs program <%+#v>", enumsDB, enumsLocal)
		log.Println(err.Error())
		return false, err

	}
	return true, nil
}

func (m *metricDBContext) isShortMetricsTableExists() (bool, error) {
	query := "select tablename from pg_catalog.pg_tables where schemaname='public' and tablename='" + m.tableName + "'"
	// row := m.conn.QueryRow(context.Background(), stmt2)
	// name := ""
	// err := row.Scan(&name)
	// if err == pgx.ErrNoRows {
	// 	log.Printf("table %s not exists in DB\n", m.tableName)
	// 	return false, nil
	// }
	// fmt.Printf("check table: err(%v),  %s\n", err, name)
	// return true, nil
	err := m.conn.QueryRow(context.Background(), query).Scan(nil)
	if err == pgx.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (m *metricDBContext) createMetricTable() error {
	createStmt := fmt.Sprintf(`
		-- add prevent update on this table .. ??
		-- ref: https://stackoverflow.com/questions/28415081/postgresql-prevent-updating-columns-everyone
		create table if not exists public.%s (
			id serial primary key, 
			metric_group %s not null,
			metric_name varchar(100) not null,
			metric_timestamp timestamp not null default current_timestamp, -- check ref: https://stackoverflow.com/questions/9556474/how-do-i-automatically-update-a-timestamp-in-postgresql
			metric_data bytea not null
		 ) ; `, m.tableName, m.groupTypeName)
	log.Printf("creat table: \n%s\n", createStmt)
	tag, err := m.conn.Exec(m.ctx, createStmt)
	if err != nil {
		log.Printf("create table failed: %s\n", err)
		return err
	}
	log.Printf("create table result: %s\n", tag)
	return nil
}

func (m *metricDBContext) createMetricGroupType() error {
	stmt := fmt.Sprintf("CREATE TYPE  %s AS ENUM (", m.groupTypeName)
	for _, mg := range ListMetricGroups() {
		stmt += fmt.Sprintf("'%s',", mg)
	}
	stmt = stmt[:len(stmt)-1]
	stmt += ");"

	log.Printf("creat enum: \n%s\n", stmt)
	tag, err := m.conn.Exec(context.Background(), stmt)
	if err != nil {
		log.Printf("create enum failed: %s\n", err)
		return err
	}
	log.Printf("create enum result: %s\n", tag)
	return err
}

func (m *metricDBContext) GetMetrics(count, ofs int) ([]*MetricDBRecord, error) {
	if !m.ready {
		return nil, errors.New("metric database is not ready")
	}
	err := m.conn.Ping(m.ctx)
	if err != nil {
		log.Printf("db ping failed: %s\n", err)
	}

	var res []*MetricDBRecord
	query := fmt.Sprintf(`SELECT * FROM %s limit %d offset %d`, m.tableName, count, ofs)
	rows, err := m.conn.Query(m.ctx, query)
	if err != nil {
		log.Printf("error getting list of metrics: %s\n", err)
		return nil, err
	}
	fields := rows.FieldDescriptions()
	log.Printf("%+#v\n", fields)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Printf("err in getting values: %s", err)
			return nil, err
		}
		log.Printf("%+#v\n", values)
		r := &MetricDBRecord{}
		for i, fn := range fields {
			switch fn.Name {
			case "id":
				r.ID = int((values[i]).(int32)) // FIXME - potential problem with large numbers as id field is serial (very big)
			case "metric_name":
				r.Metric_Name = values[i].(string)
			case "metric_data":
				r.Metric_Data = append(r.Metric_Data, values[i].([]byte)...)
			case "metric_group":
				r.Metric_Group = MetricGroupTypeFromString(values[i].(string))
			case "metric_timestamp":
				r.Metric_TimeStamp = values[i].(time.Time)
			}
		}
		res = append(res, r)
	}
	return res, err
}

// make the add metric async - otherwise we will get performance hit ...
func (m *metricDBContext) AddMetricAsync(mp MetricPacker) {
	if !m.ready {
		log.Println(errors.New("metric database is not ready"))
		return
	}
	go func() {
		if err := MDBctx.AddMetric(mp); err != nil {
			log.Printf("failed to add metric to DB: %s\n", mp.DumpObject())
		}
	}()
}

func (m *metricDBContext) AddMetric(mp MetricPacker) error {
	if !m.ready {
		return errors.New("metric database is not ready")
	}
	err := m.conn.Ping(m.ctx)
	if err != nil {
		log.Printf("db ping failed: %s\n", err)
		return err
	}
	record, err := NewDBRecord(mp)
	if err != nil {
		log.Printf("getting record failed: %s\n", err)
		return err
	}

	insertStmt := fmt.Sprintf(`
	insert into public.%s(metric_data, metric_group, metric_name) values ($1, $2, $3);`,
		m.tableName)

	tag, err := m.conn.Exec(m.ctx, insertStmt,
		record.Metric_Data, record.Metric_Group, record.Metric_Name)
	if err != nil {
		log.Printf("insert metric <%s> failed: %s\n", record.Metric_Name, err)
	} else {
		log.Printf("insert <%s> result: %s\n", record.Metric_Name, tag)
	}
	return err
}

// ------------------------------------------
type MetricDBRecord struct {
	ID               int
	Metric_Data      []byte
	Metric_Name      string
	Metric_Group     MetricGroupType
	Metric_TimeStamp time.Time
}

func NewDBRecord(mp MetricPacker) (*MetricDBRecord, error) {
	err := mp.ToMap().Error()
	if err != nil {
		log.Printf("metric convert to map failed: %s\n", err)
		return nil, err
	}
	err = mp.Encode().Error()
	if err != nil {
		log.Printf("metric encode failed: %s\n", err)
		return nil, err
	}
	err = mp.Compress().Error()
	if err != nil {
		log.Printf("metric compress failed: %s\n", err)
		return nil, err
	}
	return &MetricDBRecord{
		Metric_Name:  mp.Name(),
		Metric_Data:  mp.CompressedContent(),
		Metric_Group: mp.GroupType(),
	}, nil
}

func (r MetricDBRecord) String() string {
	return fmt.Sprintf("\tid: %d\n"+
		"\tname: %s\n"+
		"\tgroup: %s\n"+
		"\ttime: %s\n"+
		"\tdata:\n%s\n",
		r.ID, r.Metric_Name, r.Metric_Group, r.Metric_TimeStamp,
		hex.Dump(r.Metric_Data))
}

func (r MetricDBRecord) DumpMetricData() string {
	mp := NewMetricGroup(r.Metric_Group)
	err := mp.FromCompressed(r.Metric_Data).Error()
	if err != nil {
		log.Printf("err: %s\n", err)
		return ""
	}
	err = mp.Decompress().Error()
	if err != nil {
		log.Printf("err: %s\n", err)
		return ""
	}
	err = mp.Decode().Error()
	if err != nil {
		log.Printf("err: %s\n", err)
		return ""
	}
	err = mp.ToObject().Error()
	if err != nil {
		log.Printf("err: %s\n", err)
		return ""
	}
	return mp.DumpObject()
}
