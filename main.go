package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

var retryCount = 1

var timeHeuristic = time.Duration(8 * time.Second)
var longTimeHeuristic = time.Duration(30 * time.Second)
var longerTimeHeuristic = time.Duration(100 * time.Second)

var TIMEOUT = timeHeuristic

func main() {
	for i := 0; i < 3; i++ {
		if err := MakeRequest("第" + strconv.Itoa(i+1) + "次尝试"); err == nil {
			return
		} else {
			if i == 0 {
				TIMEOUT = longTimeHeuristic
			} else {
				TIMEOUT = longerTimeHeuristic
			}
		}
	}
}

func MakeRequest(count string) error {
	// start := time.Now()
	argsWithoutProg := os.Args[1:]
	log.Println(string(count))
	params, err := MakeAccountPreflightRequest()
	if err != nil {
		log.Print(err)
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下 account.ccnu.edu.cn 是否可以打开呢，错误原因：" + err.Error())
		return err
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Print(err)
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下 sentry_service 状态呢，错误原因：" + err.Error())
		return err
	}

	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
		Jar:     jar,
	}

	if err := MakeAccountRequest(argsWithoutProg[0], argsWithoutProg[1], params, &client); err != nil {
		log.Println(err.Error())
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下 account.ccnu.edu.cn 的登录呢，错误原因：" + err.Error())
		return err
	}

	if err := MakeXKRequest(&client); err != nil {
		log.Println(err.Error())
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下 xk.ccnu.edu.cn 的登录呢，错误原因：" + err.Error())
		return err
	}

	if err := MakeGradeRequest(&client); err != nil {
		log.Println(err.Error())
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下教务系统成绩查询是否正常呢，错误原因：" + err.Error())
		return err
	}

	if err := MakeTableRequest(&client); err != nil {
		log.Println(err.Error())
		SendAlert("[匣子报警][" + string(count) + "] 亲亲，这边建议您检查一下教务系统课表查询是否正常呢，错误原因：" + err.Error())
		return err
	}

	// elapsed := time.Since(start)
	// SendAlert("[华师匣子][" + string(count) + "] 亲亲，学校系统一切正常。本次请求用时：" + elapsed.String())
	return nil
}

func SendAlert(text string) {
	log.Println("发送劲爆")
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	_, err = http.Post("https://oapi.dingtalk.com/robot/send?access_token=0fc384d57235fdb1cc6dfa83408d8154507d0699f3649589dd5a7898012ee690", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Println(err)
	}
}

// account.ccnu.edu.cn 模拟登录，用于验证账号密码是否可以正常登录
func MakeAccountPreflightRequest() (*AccountReqeustParams, error) {
	var JSESSIONID string
	var lt string
	var execution string
	var _eventId string

	params := &AccountReqeustParams{}

	// 初始化 http client
	client := http.Client{
		Timeout: TIMEOUT,
	}

	// 初始化 http request
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login", nil)
	if err != nil {
		log.Println(err)
		return params, err
	}

	// 发起请求
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return params, err
	}

	// 读取 Body
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Println(err)
		return params, err
	}

	// 获取 Cookie 中的 JSESSIONID
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			JSESSIONID = cookie.Value
		}
	}

	if JSESSIONID == "" {
		log.Println("Can not get JSESSIONID")
		return params, errors.New("Can not get JSESSIONID")
	}

	// 正则匹配 HTML 返回的表单字段
	ltReg := regexp.MustCompile("name=\"lt\".+value=\"(.+)\"")
	executionReg := regexp.MustCompile("name=\"execution\".+value=\"(.+)\"")
	_eventIdReg := regexp.MustCompile("name=\"_eventId\".+value=\"(.+)\"")

	bodyStr := string(body)

	ltArr := ltReg.FindStringSubmatch(bodyStr)
	if len(ltArr) != 2 {
		log.Println("Can not get form paramater: lt")
		return params, errors.New("Can not get form paramater: lt")
	}
	lt = ltArr[1]

	execArr := executionReg.FindStringSubmatch(bodyStr)
	if len(execArr) != 2 {
		log.Println("Can not get form paramater: execution")
		return params, errors.New("Can not get form paramater: execution")
	}
	execution = execArr[1]

	_eventIdArr := _eventIdReg.FindStringSubmatch(bodyStr)
	if len(_eventIdArr) != 2 {
		log.Println("Can not get form paramater: _eventId")
		return params, errors.New("Can not get form paramater: _eventId")
	}
	_eventId = _eventIdArr[1]

	log.Println("Get params successfully", lt, execution, _eventId)

	params.lt = lt
	params.execution = execution
	params._eventId = _eventId
	params.submit = "LOGIN"
	params.JSESSIONID = JSESSIONID

	return params, nil
}

// account.ccnu.edu.cn 模拟登录，用于验证账号密码是否可以正常登录
func MakeAccountRequest(sid, password string, params *AccountReqeustParams, client *http.Client) error {
	v := url.Values{}
	v.Set("username", sid)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	request, err := http.NewRequest("POST", "https://account.ccnu.edu.cn/cas/login;jsessionid="+params.JSESSIONID, strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")

	resp, err := client.Do(request)
	if err != nil {
		log.Print(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	// check
	reg := regexp.MustCompile("class=\"success\"")
	matched := reg.MatchString(string(body))
	if !matched {
		log.Println("Wrong sid or pwd")
		return errors.New("Wrong sid or pwd")
	}

	log.Println("Login successfully")
	return nil
}

// xk.ccnu.edu.cn 模拟登录，用于请求成绩/课表等等
func MakeXKRequest(client *http.Client) error {
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fxk.ccnu.edu.cn%2Fssoserver%2Flogin%3Fywxt%3Djw%26url%3Dxtgl%2Findex_initMenu.html", nil)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = client.Do(request)
	if err != nil {
		log.Println(err)
		return err
	}

	u, err := url.Parse("http://xk.ccnu.edu.cn")
	if err != nil {
		log.Println(err)
		return err
	}

	for _, cookie := range client.Jar.Cookies(u) {
		fmt.Printf("  %s: %s\n", cookie.Name, cookie.Value)
	}
	return nil
}

// xk.ccnu.edu.cn 获取成绩
func MakeGradeRequest(client *http.Client) error {
	v := url.Values{}

	v.Set("xnm", "2017")
	v.Set("xqm", "12")
	v.Set("_search", "false")
	v.Set("nd", string(time.Now().UnixNano()))
	v.Set("queryModel.showCount", "50")
	v.Set("queryModel.currentPage", "1")
	v.Set("queryModel.sortName", "")
	v.Set("queryModel.sortOrder", "asc")
	v.Set("time", "0")

	request, err := http.NewRequest("POST", "http://xk.ccnu.edu.cn/cjcx/cjcx_cxDgXscj.html?doType=query&gnmkdm=N305005", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("Origin", "http://xk.ccnu.edu.cn")
	request.Header.Set("Host", "xk.ccnu.edu.cn")
	request.Header.Set("Referer", "http://xk.ccnu.edu.cn//cjcx/cjcx_cxDgXscj.html?gnmkdm=N305005&layout=default&su=2016210942")

	resp, err := client.Do(request)
	if err != nil {
		log.Print(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	var grade = &Grade{}

	if err := json.Unmarshal(body, &grade); err != nil {
		log.Print(err)
		return err
	}

	if len(grade.Items) == 0 {
		return errors.New("empty grade list")
	}

	return nil
}

// xk.ccnu.edu.cn 获取课表
func MakeTableRequest(client *http.Client) error {
	v := url.Values{}

	v.Set("xnm", "2018")
	v.Set("xqm", "12")

	request, err := http.NewRequest("POST", "http://xk.ccnu.edu.cn/kbcx/xskbcx_cxXsKb.html?gnmkdm=N2151", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("Origin", "http://xk.ccnu.edu.cn")
	request.Header.Set("Host", "xk.ccnu.edu.cn")
	request.Header.Set("Referer", "http://xk.ccnu.edu.cn//cjcx/cjcx_cxDgXscj.html?gnmkdm=N305005&layout=default&su=2016210942")

	resp, err := client.Do(request)
	if err != nil {
		log.Print(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	var table = &Table{}

	if err := json.Unmarshal(body, &table); err != nil {
		log.Print(err)
		return err
	}

	if len(table.KbList) == 0 {
		return errors.New("empty table list")
	}

	return nil
}
