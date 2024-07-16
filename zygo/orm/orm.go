package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"web/zygo/mylog"
)

// 数据库连接和配置
type MyDb struct {
	db     *sql.DB
	logger *mylog.Logger
	Prefix string
}
type MySession struct {
	db          *MyDb
	tx          *sql.Tx
	beginTx     bool
	tableName   string
	fieldName   []string
	placeHolder []string
	values      []any
	updateParam strings.Builder
	whereParam  strings.Builder
	whereValue  []any
}

func Open(driverName, dataSourceName string) *MyDb {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	myDb := &MyDb{db: db, logger: mylog.Default()}

	//最大空闲连接数，默认不配置，是2个最大空闲连接
	db.SetMaxIdleConns(5)
	//最大连接数，默认不配置，是不限制最大连接数
	db.SetMaxOpenConns(100)
	// 连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	//空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return myDb
}

func (db *MyDb) New(data any) *MySession {
	m := &MySession{db: db}
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}
	tVar := t.Elem()
	m.tableName = db.Prefix + strings.ToLower(Name(tVar.Name()))
	return m
}

func (s *MySession) Table(table string) *MySession {
	s.tableName = table
	return s
}

/*单个插入insert into table xxx values (?,?,?)*/
func (s *MySession) Insert(data any) (int64, int64, error) {
	s.fieldNames(data)

	query := fmt.Sprintf("insert into %s (%s) values (%s)", s.tableName, strings.Join(s.fieldName, ","), strings.Join(s.placeHolder, ","))
	s.db.logger.Info(query)
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	if err != nil {
		return -1, -1, err
	}
	exec, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}

	id, err := exec.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	s.tx.Commit()
	return id, affected, nil
}

/*批量插入insert into table xxx values (?,?,?)，(?,?,?)*/
func (s *MySession) InsertBatch(data []any) (int64, int64, error) {
	if len(data) == 0 {
		panic("data is empty")
		return -1, -1, errors.New("data is empty")
	}
	s.fieldNames(data[0])
	query := fmt.Sprintf("insert into %s (%s) values", s.tableName, strings.Join(s.fieldName, ","))
	var sb strings.Builder
	sb.WriteString(query)
	s.values = make([]any, 0)
	for index, value := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
		s.valueBatch(value)
	}

	s.db.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	if err != nil {
		return -1, -1, err
	}
	exec, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}

	id, err := exec.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	s.tx.Commit()
	return id, affected, nil
}

/*用于链式调用*/
func (s *MySession) UpdateParam(key string, value any) *MySession {
	if s.updateParam.String() != "" {
		s.updateParam.WriteString(" , ")
	}
	s.updateParam.WriteString(key)
	s.updateParam.WriteString(" = ?")
	s.values = append(s.values, value)
	return s
}
func (s *MySession) UpdateMap(m map[string]any) *MySession {
	for k, v := range m {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(" , ")
		}
		s.updateParam.WriteString(k)
		s.updateParam.WriteString(" = ?")
		s.values = append(s.values, v)
	}
	return s
}

/*更新*/
func (s *MySession) Update(data ...any) (int64, int64, error) {
	if len(data) > 2 {
		panic(errors.New("data is not allow"))
	}
	var sb strings.Builder
	var sign = true
	if len(data) == 2 {
		sign = false
	}
	//没有传参就是链式调用
	if len(data) == 0 {
		query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
		sb.WriteString(query)
		sb.WriteString(s.whereParam.String())
		s.values = append(s.values, s.whereValue...)
		s.db.logger.Info(sb.String())
		var stmt *sql.Stmt
		var err error
		if s.beginTx {
			stmt, err = s.tx.Prepare(query)
		} else {
			stmt, err = s.db.db.Prepare(query)
		}
		if err != nil {
			return -1, -1, err
		}
		exec, err := stmt.Exec(s.values...)
		if err != nil {
			return -1, -1, err
		}

		id, err := exec.LastInsertId()
		if err != nil {
			return -1, -1, err
		}
		affected, err := exec.RowsAffected()
		if err != nil {
			return -1, -1, err
		}
		s.tx.Commit()
		return id, affected, nil
	}
	if !sign {
		//update table x where id = ? set age = 1, name = 1;
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(" , ")
		}
		s.updateParam.WriteString(data[0].(string))
		s.updateParam.WriteString(" = ?")
		s.values = append(s.values, data[1])
	} else {
		//传入结构体
		updateData := data[0]
		t := reflect.TypeOf(updateData)
		v := reflect.ValueOf(updateData)
		if t.Kind() != reflect.Ptr {
			panic(errors.New("data must be a pointer"))
		}
		tVar := t.Elem()
		vVar := v.Elem()
		if s.tableName == "" {
			s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
		}
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag.Get("myorm")
			if tag == "" {
				tag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(tag, "auto_increment") {
					//自增长的主键id,在建表的时候mysql做了
					continue
				}
				if strings.Contains(tag, ",") {
					tag = tag[:strings.Index(tag, ",")]
				}

			}
			id := vVar.Field(i).Interface()
			if strings.ToLower(tag) == "id" && IsAutoId(id) {
				continue
			}
			s.updateParam.WriteString(tag)
			s.updateParam.WriteString(" = ?")
			if i < tVar.NumField()-1 {
				s.updateParam.WriteString(",")
			}
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
	s.values = append(s.values, s.whereValue...)
	query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	//sb.WriteString(";")
	s.db.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	//stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, -1, err
	}
	exec, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}

	id, err := exec.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	s.tx.Commit()
	return id, affected, nil
}

func (s *MySession) Where(key string, value any) *MySession {
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(key)
	s.whereParam.WriteString(" = ?")
	s.whereValue = append(s.whereValue, reflect.ValueOf(value).Interface())
	return s
}

// avg聚合函数
func (s *MySession) AVG(field string) (float64, error) {
	query := fmt.Sprintf("select AVG(%s) from %s", field, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, err
	}
	row := stmt.QueryRow(s.whereValue...)
	if row.Err() != nil {
		return -1, row.Err()
	}
	var result float64
	err_ := row.Scan(&result)
	if err_ != nil {
		return -1, err_
	}

	return result, nil
}

// 聚合函数总结
func (s *MySession) Aggregate(funcName string, field string) (int64, error) {
	var fieldSb strings.Builder
	fieldSb.WriteString(funcName)
	fieldSb.WriteString("(")
	fieldSb.WriteString(field)
	fieldSb.WriteString(")")

	query := fmt.Sprintf("select %s from %s", fieldSb.String(), s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, err
	}
	row := stmt.QueryRow(s.whereValue...)
	if row.Err() != nil {
		return -1, row.Err()
	}
	var result int64
	err_ := row.Scan(&result)
	if err_ != nil {
		return -1, err_
	}
	return result, nil
}

func (s *MySession) And() *MySession {
	s.whereParam.WriteString(" and ")
	return s
}

func (s *MySession) Or() *MySession {
	s.whereParam.WriteString(" or ")
	return s
}

/*单个查询方法*/
func (s *MySession) SelectOne(data any, fieldsnames []string) error {
	filename := "*"
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}
	if fieldsnames != nil {
		filename = strings.Join(fieldsnames, ",")
	}
	query := fmt.Sprintf("select %s from %s  %s", filename, s.tableName, s.whereParam.String())
	s.db.logger.Info(query)
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(s.whereValue...)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := 0; i < len(columns); i++ {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag.Get("myorm")
			if tag == "" {
				tag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(tag, ",") {
					tag = tag[:strings.Index(tag, ",")]
				}
			}

			for index, column := range columns {
				if column == tag {
					target := values[index]
					targetValue := reflect.ValueOf(target).Interface()
					fieldType := tVar.Field(i).Type
					result := reflect.ValueOf(targetValue).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}

	}
	return nil
}

func (s *MySession) Delete() (int64, error) {
	//delete from table where id = ?
	query := fmt.Sprintf("delete from %s ", s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	//stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return -1, err
	}
	r, err := stmt.Exec(s.whereValue...)
	if err != nil {
		return -1, err
	}
	s.tx.Commit()
	return r.RowsAffected()

}

/* SetMaxIdleConns 最大空闲连接数，默认不配置，是2个最大空闲连接*/
func (db *MyDb) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

/*给会话的value赋值*/
func (s *MySession) valueBatch(data any) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Ptr {
		panic(errors.New("data must be a pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag.Get("myorm")
		if tag == "" {
			tag = strings.ToLower(Name(fieldName))
		} else {
			if strings.Contains(tag, "auto_increment") {
				//自增长的主键id,在建表的时候mysql做了
				continue
			}
		}
		id := vVar.Field(i).Interface()
		if strings.ToLower(tag) == "id" && IsAutoId(id) {
			continue
		}
		s.values = append(s.values, vVar.Field(i).Interface())
	}
}

func (s *MySession) fieldNames(data any) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Ptr {
		panic(errors.New("data must be a pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	//if s.tableName == "" {
	//	s.tableName = s.db.Prefix + strings.ToLower(tVar.Name())
	//}
	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag.Get("myorm")
		if tag == "" {
			tag = strings.ToLower(Name(fieldName))
		} else {
			if strings.Contains(tag, "auto_increment") {
				//自增长的主键id,在建表的时候mysql做了
				continue
			}
			if strings.Contains(tag, ",") {
				tag = tag[:strings.Index(tag, ",")]
			}

		}
		id := vVar.Field(i).Interface()
		if strings.ToLower(tag) == "id" && IsAutoId(id) {
			continue
		}
		s.fieldName = append(s.fieldName, tag)
		s.placeHolder = append(s.placeHolder, "?")
		s.values = append(s.values, vVar.Field(i).Interface())
	}
}
func IsAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if id.(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if id.(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if id.(int) <= 0 {
			return true
		}
	}
	return false
}

/*数据库链接关闭*/
func (db *MyDb) Close() error {
	return db.db.Close()
}

// 驼峰命名转化为_格式
func Name(fieldName string) string {
	var names = fieldName[:]
	lastIndex := 0
	var sb strings.Builder
	for index, value := range names {
		if value >= 65 && value <= 90 {
			if index == 0 {
				continue
			}
			sb.WriteString(names[lastIndex:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	sb.WriteString(names[lastIndex:])
	return sb.String()
}

/*查询多个*/
func (s *MySession) Select(data any, fieldsnames []string) ([]any, error) {
	filename := "*"
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return nil, errors.New("data must be pointer")
	}
	if fieldsnames != nil {
		filename = strings.Join(fieldsnames, ",")
	}
	query := fmt.Sprintf("select %s from %s  %s", filename, s.tableName, s.whereParam.String())
	s.db.logger.Info(query)
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(s.whereValue...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]any, 0)
	for {
		if rows.Next() {
			//由于传进地址，每次赋值，实际都是一个result里面，值都一样
			//希望每次查询都能换地址
			data := reflect.New(t.Elem()).Interface()
			values := make([]any, len(columns))
			fieldScan := make([]any, len(columns))
			for i := 0; i < len(columns); i++ {
				fieldScan[i] = &values[i]
			}
			err := rows.Scan(fieldScan...)
			if err != nil {
				return nil, err
			}
			tVar := t.Elem()
			vVar := reflect.ValueOf(data).Elem()
			for i := 0; i < tVar.NumField(); i++ {
				fieldName := tVar.Field(i).Name
				tag := tVar.Field(i).Tag.Get("myorm")
				if tag == "" {
					tag = strings.ToLower(Name(fieldName))
				} else {
					if strings.Contains(tag, ",") {
						tag = tag[:strings.Index(tag, ",")]
					}
				}

				for index, column := range columns {
					if column == tag {
						target := values[index]
						targetValue := reflect.ValueOf(target).Interface()
						fieldType := tVar.Field(i).Type
						result := reflect.ValueOf(targetValue).Convert(fieldType)
						vVar.Field(i).Set(result)
					}
				}

			}
			results = append(results, data)
		} else {
			break
		}

	}
	return results, nil
}

func (s *MySession) Like(field string, value any) *MySession {
	//name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValue = append(s.whereValue, "%"+value.(string)+"%")
	return s
}
func (s *MySession) LikeRight(field string, value any) *MySession {
	//name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValue = append(s.whereValue, value.(string)+"%")
	return s
}

func (s *MySession) LikeLeft(field string, value any) *MySession {
	//name like %s
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValue = append(s.whereValue, "%"+value.(string))
	return s
}

func (s *MySession) Group(field ...string) *MySession {
	//group by aa,b,
	s.whereParam.WriteString(" group by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	return s
}

func (s *MySession) OrderDesc(field ...string) *MySession {
	//order by ss,s DESC
	s.whereParam.WriteString("order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" DESC ")
	return s
}

func (s *MySession) OrderAsc(field ...string) *MySession {
	//order by ss,s DESC
	s.whereParam.WriteString("order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" ASC ")
	return s
}

// order aa desc,bb,asc
func (s *MySession) Order(field ...string) *MySession {
	if len(field)%2 != 0 {
		panic("fields must be even")
	}
	s.whereParam.WriteString("order by ")
	for index, value := range field {
		s.whereParam.WriteString(value + " ")
		if index%2 != 0 && index < len(field)-1 {
			s.whereParam.WriteString(",")
		}
	}
	return s
}

/*原生sql*/
func (s *MySession) Exec(query string, args ...interface{}) (int64, error) {
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	//stmt, err := s.db.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	exec, err := stmt.Exec(args)
	if err != nil {

		return -1, err
	}
	if strings.Contains(strings.ToLower(query), "insert") {
		s.tx.Commit()
		return exec.LastInsertId()
	}
	s.tx.Commit()
	return exec.RowsAffected()
}

func (s *MySession) QueryRow(query string, data any, queryValue ...any) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}

	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	//stmt, err := s.db.db.Prepare(sql)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(queryValue...)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := 0; i < len(columns); i++ {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag.Get("myorm")
			if tag == "" {
				tag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(tag, ",") {
					tag = tag[:strings.Index(tag, ",")]
				}
			}

			for index, column := range columns {
				if column == tag {
					target := values[index]
					targetValue := reflect.ValueOf(target).Interface()
					fieldType := tVar.Field(i).Type
					result := reflect.ValueOf(targetValue).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}

	}
	s.tx.Commit()
	return nil
}
func (s *MySession) Begin() error {
	tx, err := s.db.db.Begin()
	if err != nil {
		return err
	}
	s.tx = tx
	s.beginTx = true
	return nil
}

func (s *MySession) Commit() error {
	err := s.tx.Commit()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

func (s *MySession) Rollback() error {
	err := s.tx.Rollback()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}
