package main

import(
    "reflect"
    "fmt"
    "time"
    "math/rand"
    "errors"
    "strings"
)
func mapToStruct(m map[string]interface{}, s interface{}) error {
    v := reflect.ValueOf(s).Elem()

    for key, value := range m {
        structField := v.FieldByName(key)
        if !structField.IsValid() {
            continue
            // return fmt.Errorf("No such field: %s in obj", key)
        }

        if !structField.CanSet() {
            continue
            // return fmt.Errorf("Cannot set %s field value", key)
        }

        val := reflect.ValueOf(value)
        if structField.Type() != val.Type() {
            return fmt.Errorf("Provided value type didn't match obj field type")
        }

        structField.Set(val)
    }
    return nil
}

func UnixToStr(unix int64) string {
	u := time.Unix(unix, 0).Format("2006-01-02 15:04:05")
	return u
}

func ByteSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%.2fB", float64(size)/float64(1))
	} else if size < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
	} else if size < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(size)/float64(1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(size)/float64(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fTB", float64(size)/float64(1024*1024*1024*1024))
	}
}

type BtnMessage struct{
    Type string
    Value interface{}
}

func packBtnMsg(ty string,value interface{})string{
    s := fmt.Sprintf("%v=%v",ty,value )
    fmt.Println(s)
    return s
}

func unPackBtnMsg(s string)(btn BtnMessage,err error){
    arr := strings.Split(s, "=")
    if len(arr)<2{
        return btn, errors.New("数据格式错误")
    }
    btn.Type=arr[0]
    btn.Value=arr[1]
    return btn,nil

}
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	fmt.Println(string(code))
	return string(code)
}