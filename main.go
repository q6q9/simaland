package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Product struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	CategoryID  int    `json:"category_id"`
	Slug        string `json:"slug"`
	Description string `json:"description""`
	Price       int    `json:"price"`
	NewPrice    int    `json:"new_price"`
	DateAdd     int    `json:"dt_add"`
	DateUpdate  int    `json:"dt_update"`
	Cover       string `json:"cover"`

	Images []struct {
		Img      string `json:"img"`
		ImgThumb string `json:"img_thumb"`
	} `json:"images"`
}
type Category struct {
	Icon            string `json:"icon"`
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Parent_id       int    `json:"parent_id"`
	Slug            string `json:"slug"`
	Status          int    `json:"status"`
	Type            int    `json:"type"`
	metaDescription string
	metaTitle       string
}

const CompanyId = 1
const TableNameProducts = "products"
const FileNameProducts = "products.json"
const TableFieldsProducts = "(cover, id, title, category_id, slug, description, price, new_price, dt_add, dt_update, status, type, company_id)"
const CurrentIdLastProducts = 2000

const CurrentIdLastCategory = 7000
const TableNameCategory = "category_shop"
const FileNameCategory = "categories.json"
const TableFieldsCategory = "(icon, id, name, parent_id, slug, status, type, meta_title, meta_description)"

const TableNameProductsImages = "products_img"
const TableFieldsProductsImages = "(img, img_thumb, product_id)"

var db *sql.DB

func main() {
	connectDB()
	importAll()
}

func connectDB() {
	cfg := mysql.Config{
		User:   "q6q9",
		Passwd: "123qweasd",
		Net:    "tcp",
		Addr:   "localhost",
		DBName: "da",
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
}

func importAll() {
	fmt.Printf(`ВНИМАНИЕ! Для продуктов:
	company_id = %d
	последний ID продукта = %d
	имя таблицы = %v 
Для продуктов:
	последний ID категории = %d
	имя таблицы = %v    
напишите "YES",чтобы продолжить:
`, CompanyId, CurrentIdLastProducts, TableNameProducts, CurrentIdLastCategory, TableNameCategory)
	var input string
	_, _ = fmt.Fscan(os.Stdin, &input)
	if input != "YES" {
		panic("Not \"YES\"")
	}

	var categories []Category
	getFileContentJSON(FileNameCategory, &categories)
	for i, category := range categories {
		sqlText, err := insertCategory(category)
		if err != nil {
			fmt.Println(err)
			fmt.Println(sqlText)
			fmt.Println()
		}
		if i%20 == 0 {
			fmt.Printf("Insert %d categories\n", i)
		}
	}

	var products []Product
	getFileContentJSON(FileNameProducts, &products)
	for i, product := range products {
		sqlText, err := insertProduct(product)
		if err != nil {
			fmt.Println(err)
			fmt.Println(sqlText)
			fmt.Println("########################")
		}
		insertProductImages(product)
		if err != nil {
			fmt.Println(err)
			fmt.Println(sqlText)
			fmt.Println("-_-_-_-_-_-_-_-_-_-_-_-_")
			fmt.Println("-_-_-_-_-_-_-_-_-_-_-_-_")
		}

		if i%20 == 0 {
			fmt.Printf("Insert %d products\n", i)
		}
	}
}

func getFileContentJSON(fileName string, items interface{}) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(content), &items)
	if err != nil {
		panic(err)
	}
}

func insertCategory(product Category) (string, error) {
	parent_id := product.Parent_id+CurrentIdLastCategory
	if product.Parent_id == 0 {
		parent_id = 0;
	}
	record := fmt.Sprintf(" ('%v', '%v', '%v', '%v', '%v', '1', '2', '', '')",
		product.Icon,
		product.Id+CurrentIdLastCategory,
		product.Name,
		parent_id,
		product.Slug)
	sqlText := fmt.Sprintf(`INSERT INTO %v %v VALUES %v`, TableNameCategory, TableFieldsCategory, record)

	_, err := db.Exec(sqlText)
	return sqlText, err
}

func insertProduct(product Product) (string, error) {
	record := fmt.Sprintf(" ('%v','%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v')",
		product.Cover,
		product.ID+CurrentIdLastProducts,
		strings.Replace(product.Title, "'", "\"", -1),
		product.CategoryID+CurrentIdLastCategory,
		strings.Replace(product.Slug, "'", "\"", -1),
		strings.Replace(product.Description, "'", "\"", -1),
		product.Price,
		product.NewPrice,
		product.DateAdd,
		product.DateUpdate,
		1,
		2,
		CompanyId)

	sqlText := fmt.Sprintf(`INSERT INTO %v %v VALUES %v`, TableNameProducts, TableFieldsProducts, record)

	_, err := db.Exec(sqlText)
	if !((err != nil) && strings.Split(fmt.Sprintf("%v", err), " ")[1][:4] != "1452") {
		err = nil
	}
	return sqlText, err
}

func insertProductImages(product Product) {
	for _, img := range product.Images {
		record := fmt.Sprintf(" ('%v','%v', '%v')", img.Img, img.ImgThumb, product.ID+CurrentIdLastProducts)

		sqlText := fmt.Sprintf(`INSERT INTO %v %v VALUES %v`,
			TableNameProductsImages, TableFieldsProductsImages, record)
		_, err := db.Exec(sqlText)
		if err != nil && strings.Split(fmt.Sprintf("%v", err), " ")[1][:4] != "1452" {
			log.Fatal(err)
		}
	}
}

//func exportAllProducts() []Product {
//	res, err := db.Query(`select slug, category_id, price, IF(new_price IS NULL, 0, new_price) from products`)
//	if err != nil {
//		panic(err)
//	}
//
//	var products = []Product{}
//
//	for res.Next() {
//		var product Product
//		err := res.Scan(&product.Slug, &product.CategoryID, &product.Price, &product.NewPrice)
//		if err != nil {
//			log.Fatal(err)
//		}
//		products = append(products, product)
//	}
//	return products
//}
