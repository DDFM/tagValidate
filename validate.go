package tagValidate

import (
	"strings"
	"errors"
	"fmt"
	"reflect"
)

// 验证结果返回类型.
type ValidateResult struct {
	ErrorMgs map[string][]string
	IsSuccess bool 
	ErrorCode int
	Msg map[string][]string
}

// 通过struct中tag验证数据规范.
// e.g.:`dd_sf_v:"{m:'len(0,2)',msg:'长度不合法',type:'2',mode:'pass'}||{m:'empty()',msg:'',type:'',mode:'notpass'}"`.
// e.g.:`dd_sf_v:"len()"`.
// e.g.:`dd_sf_v:{m:'len()'}`.
// 循环引用问题暂未处理.
func Validate(i interface{},result *ValidateResult) (bool, error) {
	if i == nil {
		return false, errors.New("i is nil")
	}
	if result==nil{
		result=&ValidateResult{}
	}
	result.IsSuccess=true
	result.ErrorMgs=make(map[string][]string)
	result.Msg=make(map[string][]string)
	return validate("i", reflect.ValueOf(i),result)
}

// 验证数据.
func validate(path string, v reflect.Value,result *ValidateResult) (bool, error) {	
	temppath := ""
	switch v.Kind() {
	case reflect.Invalid:
		{
			return false, errors.New("v's type is Invalid")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.String:
		{
			fmt.Printf("%s=%v\n", path, v.Interface())
		}
	case reflect.Slice, reflect.Array:
		{
			for i := 0; i < v.Len(); i++ {
				temppath = fmt.Sprintf("%s[%d]", path, i)
				validate(temppath, v.Index(i),result)
			}
		}
	case reflect.Struct:
		{
			for i := 0; i < v.NumField(); i++ {
				temppath = fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
				validateTag(temppath,v.Type().Field(i),v.Field(i),result)
				validate(temppath, v.Field(i),result)
			}
		}
	case reflect.Map:
		{
			for _, key := range v.MapKeys() {
				temppath = fmt.Sprintf("%s[%s]", path, key)
				validate(temppath, v.MapIndex(key),result)
			}
		}
	case reflect.Ptr:
		{
			if v.IsNil() {
				fmt.Printf("%s=nil\n", path)
			} else {
				temppath = fmt.Sprintf("(*%s)", path)
				validate(temppath, v.Elem(),result)
			}
		}
	case reflect.Interface:
		{
			if v.IsNil() {
				fmt.Printf("%s=nil\n", path)
			} else {
				temppath = path + ".value"
				validate(temppath, v.Elem(),result)
			}
		}
	default:
		{			
		}
	}
	return result.IsSuccess, nil

}

// tag标签信息结构.
type TagInfo struct{
	M string
	Msg string 
	Type string
	Mode string	
	V reflect.Value
	Path string
}

// 验证函数类型.
type FuncValidate func(TagInfo)(bool,string)

//为空验证
var empty=func(taginfo TagInfo)(bool,string){
	
	
	return false,"为空"
}

// tag map.
// key 对应m值.
// value 对应进行校验的函数.
var TagMap= map[string]FuncValidate{
	"empty":empty,
}

//查找tag标签进行校验.
func validateTag(path string, sf reflect.StructField,v reflect.Value,result *ValidateResult) {
	if strtag:=sf.Tag.Get("dd_sf_v");len(strtag)!=0{
		strTags:=make([]string,0)
		tags:=make([]TagInfo,0)
		//结构:{}||{}
		if strtag[0]=='{'&&strtag[len(strtag)-1]=='}'{	
			for _,j:=range strings.Split(strtag,"||"){			
				j=j[1:len(j)-1]
				strTags=append(strTags,j)				
			}
		}else {
			strTags=append(strTags,strtag)
		}
		if len(strTags)>0{
			for _,t:=range strTags{
			for _,a:=range strings.Split(t,","){
				tag:=TagInfo{}
				 str:=strings.Split(a,":")
				if len(str)==1{
					tag.M=str[0]
				}else {
					str[1]=strings.Trim(str[1],"'")
					switch strings.ToLower(str[0]){
					case "m":{
						tag.M=str[1]
					}
					case "msg":{
						tag.Msg=str[1]
					}
					case "type":{
						tag.Type=str[1]
					}
					case "mode":{
						tag.Mode=str[1]
					}
					}
				}
				tag.Path=path
				tag.V=v
				tags=append(tags,tag)
			}}
		}
		fmt.Print("\n----")
		fmt.Printf("%v",tags)
		fmt.Print("\n")
		//验证数据合法性
		for _,tag:=range tags{			
			if TagMap[tag.M]!=nil{
				if success,msg:=TagMap[tag.M](tag);!success{
					result.IsSuccess=false
					result.ErrorMgs[path]=append(result.ErrorMgs[path],msg)
					result.Msg[path]=append(result.Msg[path],msg)
				}
			}
		}
	}	
}
