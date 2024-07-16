package service

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"web/zygo/orm"
)

type User struct {
	Id       int64
	UserName string
	Password string
	Age      int
}

/*单个插入*/
func SaveUser() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{
		UserName: "mszlu",
		Password: "123456",
		Age:      30,
	}
	id, _, err := db.New(&User{}).Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

/*批量插入*/
func SaveUserBatch() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{
		UserName: "zy",
		Password: "123456",
		Age:      20,
	}

	user1 := &User{
		UserName: "sy",
		Password: "123456",
		Age:      18,
	}
	users := make([]any, 0)
	users = append(users, user, user1)
	id, _, err := db.New(&User{}).InsertBatch(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

/*更新*/
func UpdateUserByUser() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user1 := &User{
		UserName: "sy",
		Password: "123456",
		Age:      18,
	}
	id, _, err := db.New(&User{}).Where("id", 9).Update(user1)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func UpdateUser() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"

	id, _, err := db.New(&User{}).Table("zygo_user").Where("id", 9).Update("age", 17)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

/*链式调用*/
func UpdateUserByMethod() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"

	//id, _, err := db.New().Table("zygo_user").Where("id", 9).UpdateParam("age", 17).UpdateParam("user_name", "ssh").Update()

	m := make(map[string]interface{})
	m["user_name"] = "sh"
	m["age"] = 189
	id, _, err := db.New(&User{}).Table("zygo_user").Where("id", 9).UpdateMap(m).Update()

	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

/*查询一个*/
func SelectOne() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	err := db.New(user).Where("id", 1).SelectOne(user, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(user)

	db.Close()
}
func Where() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	err := db.New(user).Where("id", 10).And().Where("age", 20).SelectOne(user, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(user)

	db.Close()
}

func Delete() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	affecte, err := db.New(user).Where("id", 10).And().Where("age", 20).Delete()
	if err != nil {
		panic(err)
	}
	fmt.Println(affecte)
	db.Close()
}

func Select() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	results, err := db.New(user).OrderAsc("age").Select(user, nil)
	if err != nil {
		panic(err)
	}
	for _, v := range results {
		u := v.(*User)
		fmt.Println(u)
	}
	db.Close()
}

func Avg() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	results, err := db.New(user).AVG("age")
	if err != nil {
		panic(err)
	}
	fmt.Println(results)
	db.Close()
}

func Aggregate() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	user := &User{}
	results, err := db.New(user).Aggregate("count", "*")
	if err != nil {
		panic(err)
	}
	fmt.Println(results)
	db.Close()
}

func QueryRow() {
	dataSourceName := fmt.Sprintf("root:123@tcp(localhost:3306)/zygo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dataSourceName)
	db.Prefix = "zygo_"
	var user User
	sql := "SELECT * FROM zygo_user WHERE age = ?"
	err := db.New(&user).QueryRow(sql, &user, 30)
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
	db.Close()
}
