package main

/*
Go のインタフェース設計

参考
https://code.google.com/p/go-wiki/wiki/GoForCPPProgrammers
http://jordanorelli.tumblr.com/post/32665860244/how-to-use-interfaces-in-go
http://research.swtch.com/interfaces
http://www.airs.com/blog/archives/277
http://stackoverflow.com/questions/6372474/in-golang-how-to-determine-an-interface-values-real-type
http://d.hatena.ne.jp/repeatedly/20101110/1289320794
http://www.softwareresearch.net/fileadmin/src/docs/teaching/SS10/Sem/Paper__aigner_baumgartner.pdf
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// 基本的な Struct
type Point struct {
	X int
	Y int
}

// Struct のメソッド
func (p Point) Coordinate() string {
	// p がレシーバ
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main1() {
	var a Point = Point{2, 3}
	fmt.Println(a.Coordinate()) // (2, 3)
}

////////////////////////////

// 既存の型を元にした独自の型を定義できる。
type trimmedString string

// そこにメソッドも追加できる。
func (t trimmedString) trim() trimmedString {
	return t[:3]
}

func main2() {
	var t trimmedString = "abcdefg"
	fmt.Println(t.trim())

	// 型変換
	var s string = string(t)

	// 型を変換したので、 trim() は無い
	// fmt.Println(s.trim())
	fmt.Println(s)
}

////////////////////////////

// Interface を宣言
type Accessor interface {
	GetText() string
	SetText(string)
}

// Accessor を満たす実装
// Interface の持つメソッド群を実装していれば、
// Interface を満たす(satisfy) といえる。
// 明示的な宣言は必要なく、実装と完全に分離している。
type Document struct {
	text string
}

func (d *Document) GetText() string {
	return d.text
}

func (d *Document) SetText(text string) {
	d.text = text
}

func main3() {
	// Document のインスタンスを直接変更しても
	// 値渡しになってしまうので
	// ポインタを使用
	var doc *Document = &Document{}
	doc.SetText("document")
	fmt.Println(doc.GetText())

	// Accessor Interface を実装しているので
	// Accessor 型に代入可能
	var acsr Accessor = &Document{}
	acsr.SetText("accessor")
	fmt.Println(acsr.GetText())
}

////////////////////////////

type Page struct {
	Document // 匿名型を含むと、その型のメソッドが継承(というか mixin)される
	Page     int
}

func main4() {
	// Page は Document を継承しており
	// Accessor Interface を満たす。
	// この場合代入可能
	var acsr Accessor = &Page{}
	// この値は acsr.Document.text に設定されてる。
	// acsr の構造体がレシーバになっているわけではないということ
	acsr.SetText("page")
	fmt.Println(acsr.GetText())

	// Document と Page の間に代入可能な関係は無い
	// var page Page = Document{}
	// var doc Document = Page{}
}

////////////////////////////

/*
	Duck Typing
	アヒルのように鳴くなら、それはアヒル。
	動的な型付け言語では、「鳴く」ことができるかは
	実行時に試すか、予めメソッドの先頭などで調べる必要がある。

	Go の場合は「鳴く」こと自体を interface で定義し、
	コンパイル時にチェックできる。

	Accessor をみたいしていれば、 Get, Set できるという例。
*/

func SetAndGet(acsr Accessor) {
	acsr.SetText("accessor")
	fmt.Println(acsr.GetText())
}

func main5() {
	// どちらも Accessor として振る舞える
	SetAndGet(&Page{})
	SetAndGet(&Document{})
}

////////////////////////////

/*
	Override
*/
type ExtendedPage struct {
	Document
	Page int
}

// Document.GetText() のオーバーライド
func (ep *ExtendedPage) GetText() string {
	// int -> string は strconv.Itoa 使用
	return strconv.Itoa(ep.Page) + " : " + ep.Document.GetText()
}

func main6() {
	// Accessor を実装している
	var acsr Accessor = &ExtendedPage{
		Document{},
		2,
	}
	acsr.SetText("page")
	fmt.Println(acsr.GetText()) // 2 : page
}

////////////////////////////

// Interface 型
/*
	Interface 型はメソッドを持たない Interface である。
	つまり、メソッドを持たない struct を含めて、
	全ての struct は Interface 型を満たすと言える。

	例えば、下記のメソッドは全ての型の値を受け取ることができる。

	func f(v interface {}) {
		// v
	}

	しかし、ここで重要なのは v の型は interface{} 型であるということ。
	Go の runtime では、全ての値は必ず一つの型を持つ。
	型が不定といったことはなく、メソッドの引数などでは
	可能であれば型の変換が行われる。
	上記 f() は全ての値を interface{} 型に変換する。

	ある interface を実装していることを実行時にランタイムで調べるためには、
	型アサーション(type assertion) を使う。

	v := value.(typename)

	参考 http://golang.org/doc/effective_go.html#interface_conversions
*/

// Get() があるかを調べる
// er を付ける命名が慣習
type Getter interface {
	GetText() string
}

func dynamicIf(v interface{}) string {
	// v は Interface 型

	var result string
	g, ok := v.(Getter) // v が Get() を実装しているか調べる
	if ok {
		result = g.GetText()
	} else {
		result = "not implemented"
	}
	return result
}

func dynamicSwitch(v interface{}) string {
	// v は Interface 型

	var result string

	// v が実装している型でスイッチする
	switch checked := v.(type) {
	case Getter:
		result = checked.GetText()
	case string:
		result = "not implemented"
	}
	return result
}

func main7() {
	var ep *ExtendedPage = &ExtendedPage{
		Document{},
		3,
	}
	ep.SetText("page")

	// do は Interface 型を取り
	// ジェネリクス的なことができる
	fmt.Println(dynamicIf(ep))       // 3 : page
	fmt.Println(dynamicIf("string")) // not implemented

	// 型スイッチを使う場合
	fmt.Println(dynamicSwitch(ep))       // 3 : page
	fmt.Println(dynamicSwitch("string")) // not implemented
}

////////////////////////////

// 全ての型を許容するインタフェースのようなものを作っておく
type Any interface{}

// ジェネリクス的な
type GetValuer interface {
	GetValue() Any
}

// Any 型で実装
type Value struct {
	v Any
}

// GetValuer を実装
func (v *Value) GetValue() Any {
	return v.v
}

func main8() {
	// インタフェースで受け取る
	var i GetValuer = &Value{10}
	var s GetValuer = &Value{"vvv"}

	// インタフェース型のコレクションに格納
	var values []GetValuer = []GetValuer{i, s}

	// それぞれ GetValue() が Any で呼べる
	for _, val := range values {
		fmt.Println(val.GetValue())
	}
}

////////////////////////////

/*
	interface の値は二つのポインタから成る。
	- 元になる型の、メソッドテーブル
	- 元の値が持つ値
	これがわかっていれば、下記が間違っていることがわかる。
	[]string のメソッドテーブルと値は持てるが、
	その中の値を interface には変換できないから。

	参考 http://golang.org/doc/faq#convert_slice_of_interface
*/
func PrintAll(vals []interface{}) {
	for _, val := range vals {
		fmt.Println(val)
	}
}

func main9() {
	names := []string{"one", "two", "three"}

	// これは間違い
	// PrintAll(names)

	// 明示的に変換が必要
	vals := make([]interface{}, len(names))
	for i, v := range names {
		vals[i] = v
	}
	PrintAll(vals)
}

////////////////////////////

/*
	interface の設計例
	http://jordanorelli.tumblr.com/post/32665860244/how-to-use-interfaces-in-go
	ここに二つの例があるので、それを借用する。
*/

// 1, twitter API から Time のパース

/*
	Twitter の JSON を map にパースする。
	twitter の JSON には、時間が Ruby フォーマットの文字列で格納されているので、
	それを考慮して型を考える。
	"Thu May 31 00:00:01 +0000 2012"
*/

var JSONString = `{ "created_at": "Thu May 31 00:00:01 +0000 2012" } `

func main10() {
	// map として、 {string: interface{}} としてしまえば
	// value がなんであれパースは可能
	var parsedMap map[string]interface{}

	if err := json.Unmarshal([]byte(JSONString), &parsedMap); err != nil {
		panic(err)
	}

	fmt.Println(parsedMap) // map[created_at:Thu May 31 00:00:01 +0000 2012]
	for k, v := range parsedMap {
		fmt.Println(k, reflect.TypeOf(v)) // created_at string
	}
}

////////////////////////////

/*
	本来 Go の time.Time 型であると望ましい。
	が Ruby の文字列フォーマットが time.Time の文字列とデフォルトフォーマットが違う。
	そこで、 time.Time を元に新たな型を定義する。

	また、 JSON の Unmarshaller のインタフェースは下記の様になっているので、
	UnmershalJSON() を実装すれば、 Unmarshaller は満たされる。

	type Unmarshaler interface {
		UnmarshalJSON([]byte) error
	}


	なんとデフォルトで time.RubyDate が用意されているのでそれを使う。
	http://golang.org/pkg/time/#pkg-constants (これぞ Battery Included !)
*/

type Timestamp time.Time

// Unmarshaller を実装
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	v, err := time.Parse(time.RubyDate, string(b[1:len(b)-1]))
	if err != nil {
		return err
	}
	*t = Timestamp(v)
	return nil
}

func main11() {
	var val map[string]Timestamp // 定義した型を使う

	if err := json.Unmarshal([]byte(JSONString), &val); err != nil {
		panic(err)
	}

	// パースされていることを確認
	for k, v := range val {
		fmt.Println(k, time.Time(v), reflect.TypeOf(v))
		// created_at 2012-05-31 00:00:01 +0000 +0000 main.Timestamp
	}
}

////////////////////////////

/*
	HTTP リクエストから JSON を取得し、オブジェクトにパースする。

	単純にシグネチャを考えると以下のようになる。

	GetEntity(*http.Request) (interface{}, error)

	これは、戻り値の方に汎用性を持たせて、どのような型のデータも取り出せるようにしている。
	しかし、これだと戻り値は毎回型変換しないといけないし、 Postel の法則に反する。
	(「送信するものに関しては厳密に、受信するものに関しては寛容に」)

	しかし、例えば取り出す型を User として下記のようにシグネチャを変更すると、
	型の数だけ GetXXXX が必要になる。

	GetUser(*http.Request) (User, error)

	そこでインタフェースを導入する。

*/

// 各型が、自身のパース実装を持てばよいので、そのメソッドだけ定義しておく。
type Entity interface {
	UnmarshallJSON([]byte) error
}

func GetEntity(b []byte, e Entity) error {
	// 各実装に処理を移譲
	return e.UnmarshallJSON(b)
}

// 型を定義
// User に関する必要なデータだけ取りたい型的な
type UserData struct {
	Id        int
	Name      string
	Time_Zone string
	Lang      string
}

// *_count だけ適当に取りたい型的な
type CountData struct {
	Followers_count  int
	Friends_count    int
	Listed_count     int
	Favourites_count int
	Statuses_count   int
}

// Entity を実装
// ここでは、 json モジュールになげるだけで
// 同じ実装でできてしまったが、
// 本来 Entity ごとに違う実装になる。
func (d *UserData) UnmarshallJSON(b []byte) error {
	err := json.Unmarshal(b, d)
	if err != nil {
		return err
	}
	return nil
}

func (d *CountData) UnmarshallJSON(b []byte) error {
	err := json.Unmarshal(b, d)
	if err != nil {
		return err
	}
	return nil
}

func main12() {
	// 対象の JSON 文字列
	EntityString := `{
		"id":51442629,
		"name":"Jxck",
		"followers_count":1620,
		"friends_count":617,
		"listed_count":204,
		"favourites_count":2895,
		"time_zone":"Tokyo",
		"statuses_count":17387,
		"lang":"ja"
	}`
	userData := &UserData{}
	countData := &CountData{}
	GetEntity([]byte(EntityString), userData)
	GetEntity([]byte(EntityString), countData)
	fmt.Println(*userData)  // {51442629 Jxck Tokyo ja}
	fmt.Println(*countData) // {1620 617 204 2895 17387}
}

// タグ付きの Struct を定義
type TaggedStruct struct {
	field string `tag:"tag1"`
}

func main13() {
	// reflect でタグを取得
	var ts = TaggedStruct{}
	var t reflect.Type = reflect.TypeOf(ts)
	var f reflect.StructField = t.Field(0)
	var tag reflect.StructTag = f.Tag
	var val string = tag.Get("json")
	fmt.Println(tag, val) // json:"emp_name" emp_name
}

// JSON をマッピングするために
// キー名のタグ付をつけた Struct を定義
type Employee struct {
	Name  string `json:"emp_name"`
	Email string `json:"emp_email"`
	Dept  string `json:"dept"`
}

func main14() {
	// フィールド名が Struct の Filed 名と違う JSON も
	// json:"fieldname" の形でタグを付けてあるので
	// マッピングすることができる。
	var jsonString = []byte(`
	{
		"emp_name": "john",
		"emp_email" :"john@golang.com",
		"dept" :"HR"
	}`)

	var john Employee
	err := json.Unmarshal(jsonString, &john)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", john) // {Name:john Email:john@golang.com Dept:HR}
}

func main() {
	fmt.Println(">--main1------------<")
	main1()
	fmt.Println(">--main2------------<")
	main2()
	fmt.Println(">--main3------------<")
	main3()
	fmt.Println(">--main4------------<")
	main4()
	fmt.Println(">--main5------------<")
	main5()
	fmt.Println(">--main6------------<")
	main6()
	fmt.Println(">--main7------------<")
	main7()
	fmt.Println(">--main8------------<")
	main8()
	fmt.Println(">--main9------------<")
	main9()
	fmt.Println(">--main10------------<")
	main10()
	fmt.Println(">--main11------------<")
	main11()
	fmt.Println(">--main12------------<")
	main12()
	fmt.Println(">--main13------------<")
	main13()
	fmt.Println(">--main14------------<")
	main14()
}
