package main

import (
	"LBPDumpSearch/pkg/model"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db1, err := gorm.Open(sqlite.Open("/Users/henry/Desktop/dry.db"))
	if err != nil {
		panic(err)
	}

	db2, err := gorm.Open(postgres.Open("host=localhost user=lbpsearch password=lbpsearch dbname=lbpsearch port=5432 sslmode=disable"))
	if err != nil {
		panic(err)
	}

	err = db2.AutoMigrate(&model.User{})
	if err != nil {
		panic(err)
	}

	rows, err := db1.Model(&model.User{}).Rows()
	defer rows.Close()

	c := 1
	for rows.Next() {
		var user model.User
		db1.ScanRows(rows, &user)
		fmt.Printf("Processing user #%d \"%s\"\n", c, user.NpHandle)
		c++
		db2.Create(&user)
	}
}
