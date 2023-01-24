package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/metrics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var greeting string
	err = conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	stmt := "select tablename from pg_catalog.pg_tables where schemaname='public'"
	resultTableName := ""
	queryResult := conn.QueryRow(context.Background(), stmt)
	if err = queryResult.Scan(&resultTableName); err != nil && err != pgx.ErrNoRows {
		fmt.Printf("err in stmt <%s>:\n\t%s\n", stmt, err)
		os.Exit(1)
	}
	if err == pgx.ErrNoRows {
		fmt.Printf("no tables in db\n")
	} else {
		fmt.Printf("tables on db: %s\n", resultTableName)
	}

	//
	IsShortMetricsTableExists := func(metricTableName string) bool {
		var (
			METRICS_TABLE_NAME  = metricTableName
			tablenameFieldIndex int
			tablenameFieldValue string
			metricsTableExists  bool
		)
		stmt2 := "select * from pg_catalog.pg_tables where schemaname='public'"
		queryResult2, err := conn.Query(context.Background(), stmt2)
		if err != nil && err != pgx.ErrNoRows {
			fmt.Printf("err in stmt <%s>:\n\t%s\n", stmt, err)
			os.Exit(1)
		}
		if err == pgx.ErrNoRows {
			fmt.Printf("no tables in db\n")
		} else {
			for fi, fn := range queryResult2.FieldDescriptions() {
				if strings.ToLower(fn.Name) == "tablename" {
					tablenameFieldIndex = fi
				}
				fmt.Printf("reuslt field: %d:%s\n", fi, fn.Name)
			}
			// fmt.Printf("tables on db: %s\n", resultTableName)
			for queryResult2.Next() {
				v, err := queryResult2.Values()
				if err != nil {
					panic(err)
				}
				for vi, vn := range v {
					var val string
					switch vnt := vn.(type) {
					case string:
						val = vnt
					case bool:
						val = fmt.Sprintf("%v", vnt)
					default:
						val = "value problem"
					}
					fmt.Printf("%d (%s) : %s\n", vi, queryResult2.FieldDescriptions()[vi].Name, val)
				}
				if len(v) >= tablenameFieldIndex {
					var ok bool
					tablenameFieldValue, ok = v[tablenameFieldIndex].(string)
					if ok {
						if strings.ToLower(tablenameFieldValue) == METRICS_TABLE_NAME {
							metricsTableExists = true
						}
					}
				}
				// fmt.Printf("row result: %s\n", queryResult2.FieldDescriptions())
			}
		}
		return metricsTableExists
	}
	CreateShortMetricsTable := func(tableName string) error {
		var err error
		if !CheckMetricGroupEnumExists(conn, DB_METRIC_GROUP_TYPE_ENUM_NAME) {
			// TOOD: what todo in case mismatch on enum values
			if err = CreateGroupTypeEnum(conn, DB_METRIC_GROUP_TYPE_ENUM_NAME); err != nil {
				fmt.Printf("create db group type enum failed: %s\n", err)
				return err
			}
		}
		return CreateShortMetricTable(conn, tableName, DB_METRIC_GROUP_TYPE_ENUM_NAME)
	}

	if !IsShortMetricsTableExists(DB_METRIC_TABLE_NAME) {
		fmt.Printf("short metrics table not exists in db, creating...\n")
		CreateShortMetricsTable(DB_METRIC_TABLE_NAME)
	}

	res, err := GetMetrics(conn, DB_METRIC_TABLE_NAME, 100, 0)
	if err != nil {
		fmt.Printf("problem getting metrics: %s\n", err)
	} else {
		fmt.Printf("got %d metrics:\n\n", len(res))
		for _, e := range res {
			fmt.Printf("%s\ndecoded:\n%s\n", e, e.DumpMetricData())
		}
	}

	x := insertMetric()
	fmt.Printf("dump: %s\n", x)
	insertRecord(conn, x)
}

const (
	DB_METRIC_GROUP_TYPE_ENUM_NAME = "metric_group_type_enum"
	DB_METRIC_TABLE_NAME           = "short_metrics"
)

func CheckMetricGroupEnumExists(conn *pgx.Conn, dbEnumTypeName string) bool {
	/*
	   select n.nspname as enum_schema,
	          t.typname as enum_name,
	          e.enumlabel as enum_value
	   from pg_type t
	      join pg_catalog.pg_enum e on t.oid = e.enumtypid
	      join pg_catalog.pg_namespace n ON n.oid = t.typnamespace
	   where
	   	t.typname = 'metricgroup';

	*/
	field := "enum_value"
	stmt := fmt.Sprintf(`select
	n.nspname as enum_schema,
	t.typname as enum_name,
	e.enumlabel as %s
from pg_type t
	join pg_catalog.pg_enum e on t.oid = e.enumtypid
	join pg_catalog.pg_namespace n ON n.oid = t.typnamespace
where
 t.typname = '%s';`,
		field,
		dbEnumTypeName)
	fmt.Printf("stmt to execute:\n[%s]\n", stmt)
	rows, err := conn.Query(context.Background(), stmt)
	if err != nil && err != pgx.ErrNoRows {
		fmt.Printf("err in stmt <%s>:\n\t%s\n", stmt, err)
		os.Exit(1)
	}
	if err == pgx.ErrNoRows {
		fmt.Printf("no type enum in db\n")
		return false
	} else {
		fmt.Printf("enum type on db: %s\n", rows.CommandTag())
	}
	r := map[string]bool{}
	for _, mg := range metrics.ListMetricGroups() {
		r[strings.ToLower(mg.String())] = true
	}
	var enumValue string
	var enumValueIndex int
	for fi, fn := range rows.FieldDescriptions() {
		if strings.ToLower(fn.Name) == field {
			enumValueIndex = fi
		}
		fmt.Printf("reuslt field: %d:%s\n", fi, fn.Name)
	}
	for rows.Next() {
		v, err := rows.Values()
		if err != nil {
			panic(err)
		}
		for vi, vn := range v {
			str, ok := vn.(string)
			if !ok {
				str = "value problem"
			}
			fmt.Printf("%d (%s) : %s\n", vi, rows.FieldDescriptions()[vi].Name, str)
		}
		if len(v) >= enumValueIndex {
			var ok bool
			enumValue, ok = v[enumValueIndex].(string)
			if ok {
				if !r[strings.ToLower(enumValue)] {
					fmt.Printf("value: %s not found in valid metric group types\n", enumValue)
					return false
				}
				delete(r, strings.ToLower(enumValue))
			}
		} else {
			fmt.Printf("problem with number of fields return for query (len: %d, field index %d)\n",
				len(v), enumValueIndex)
			return false
		}
		// fmt.Printf("row result: %s\n", queryResult2.FieldDescriptions())
	}
	if len(r) == len(metrics.ListMetricGroups()) {
		fmt.Printf("no metric group type found in db\n")
		return false
	}
	if len(r) > 0 {
		fmt.Printf("some types not in db: leftovers %+#v\n", r)
		return false
	}

	return true
}

func CreateGroupTypeEnum(conn *pgx.Conn, dbEnumTypeName string) error {
	/*not null
	CREATE TYPE  metricGroup AS ENUM (
	'MetricGroupGlobal',
	'MetricGroupShortCreationFailure',
	'MetricGroupShortCreationSuccess',
	'MetricGroupShortAccessInvalid',
	'MetricGroupShortAccessSuccess',
	'MetricGroupFailedServedPath',
	'MetricGroupServedPath');
	*/
	stmt := fmt.Sprintf("CREATE TYPE  %s AS ENUM (", dbEnumTypeName)
	// for mg := metrics.MetricGroupFirst + 1; mg < metrics.MetricGroupLast; mg++ {
	// 	r += fmt.Sprintf("'%s',", mg)
	// }
	for _, mg := range metrics.ListMetricGroups() {
		stmt += fmt.Sprintf("'%s',", mg)
	}
	stmt = stmt[:len(stmt)-1]
	stmt += ");"

	fmt.Printf("creat enum: \n%s\n", stmt)
	tag, err := conn.Exec(context.Background(), stmt)
	if err != nil {
		fmt.Printf("create enum failed: %s\n", err)
		return err
	}
	fmt.Printf("create enum result: %s\n", tag)
	return err
}

func CreateShortMetricTable(conn *pgx.Conn, dbTableName, dbEbumTypeName string) error {
	createStmt := fmt.Sprintf(`
		-- add prevent update on this table .. ??
		-- ref: https://stackoverflow.com/questions/28415081/postgresql-prevent-updating-columns-everyone
		create table if not exists public.%s (
			id serial primary key, 
			metric_data bytea not null,
			metric_group %s not null,
			metric_name varchar(100) not null,
			metric_timestamp timestamp not null default current_timestamp -- check ref: https://stackoverflow.com/questions/9556474/how-do-i-automatically-update-a-timestamp-in-postgresql
		 ) ; `, dbTableName, dbEbumTypeName)
	fmt.Printf("creat table: \n%s\n", createStmt)
	tag, err := conn.Exec(context.Background(), createStmt)
	if err != nil {
		fmt.Printf("create table failed: %s\n", err)
		return err
	}
	fmt.Printf("create table result: %s\n", tag)
	return err
}

type MetricDBRecord struct {
	ID               int
	Metric_Data      []byte
	Metric_Name      string
	Metric_Group     metrics.MetricGroupType
	Metric_TimeStamp time.Time
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
	mp := metrics.NewMetricGroup(r.Metric_Group)
	mp.FromCompressed(r.Metric_Data)
	fmt.Printf("err: %s\n", mp.Error())
	mp.Decompress()
	fmt.Printf("err: %s\n", mp.Error())
	mp.Decode()
	fmt.Printf("err: %s\n", mp.Error())
	mp.ToObject()
	return mp.DumpObject()
}

func GetMetrics(conn *pgx.Conn, tableName string, count, ofs int) ([]*MetricDBRecord, error) {
	var res []*MetricDBRecord
	// query := fmt.Sprintf(`SELECT id, metric_data, metric_name, metric_group, metric_timestamp FROM %s limit %d offset %d`, tableName, count, ofs)
	query := fmt.Sprintf(`SELECT * FROM %s limit %d offset %d`, tableName, count, ofs)
	rows, err := conn.Query(context.Background(), query)
	// err := pgxscan.Select(context.Background(), conn, &res, query)
	if err != nil {
		fmt.Printf("error getting list of metrics: %s\n", err)
		return nil, err
	}
	fields := rows.FieldDescriptions()
	// fmt.Printf("%+#v\n", fields)
	for rows.Next() {
		values, _ := rows.Values()
		// fmt.Printf("%+#v\n", values)
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
				r.Metric_Group = metrics.MetricGroupTypeFromString(values[i].(string))
			case "metric_timestamp":
				r.Metric_TimeStamp, _ = values[i].(time.Time)
			}
		}
		res = append(res, r)
	}
	return res, err
}

func insertMetric() *MetricDBRecord {
	m := metrics.NewMetricShortAccessInvalid()
	m.InvalidShortAccessName = uuid.NewString()
	m.InvalidShortAccessIP = "2001::3213:2313:3213"
	m.InvalidShortAccessTime = time.Now().String()
	m.InvalidShortAccessReferrer = "someone referred me to here"
	m.InvalidShortAccessInfo = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Vestibulum lorem sed risus ultricies. Rhoncus dolor purus non enim praesent elementum facilisis leo. Adipiscing vitae proin sagittis nisl rhoncus mattis rhoncus. Nunc sed id semper risus in hendrerit gravida rutrum quisque. Nibh sed pulvinar proin gravida hendrerit lectus a. Volutpat lacus laoreet non curabitur. At lectus urna duis convallis convallis tellus id interdum velit. Purus gravida quis blandit turpis. Risus ultricies tristique nulla aliquet enim tortor at. Arcu dui vivamus arcu felis bibendum ut. Turpis egestas maecenas pharetra convallis posuere. Sem nulla pharetra diam sit. Placerat duis ultricies lacus sed turpis. Vel turpis nunc eget lorem dolor sed viverra. Nisi lacus sed viverra tellus in hac habitasse platea dictumst. Volutpat diam ut venenatis tellus in."

	mp := metrics.MetricPacker(m)
	fmt.Printf("%s\n", mp.DumpObject())
	mp.ToMap().Encode()
	mp.Compress()
	return &MetricDBRecord{
		Metric_Name:  m.InvalidShortAccessName,
		Metric_Group: mp.GroupType(),
		Metric_Data:  mp.CompressedContent(),
	}
}

func insertRecord(conn *pgx.Conn, record *MetricDBRecord) error {

	insertStmt := fmt.Sprintf(`
	insert into public.%s(metric_data, metric_group, metric_name) values ($1, $2, $3);`,
		DB_METRIC_TABLE_NAME)

	tag, err := conn.Exec(context.Background(), insertStmt,
		record.Metric_Data, record.Metric_Group, record.Metric_Name)
	if err != nil {
		fmt.Printf("insert metric <%s> failed: %s\n", record.Metric_Name, err)
	} else {
		fmt.Printf("insert <%s> result: %s\n", record.Metric_Name, tag)
	}
	return err
}