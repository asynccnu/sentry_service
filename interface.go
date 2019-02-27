package main

type AccountReqeustParams struct {
	lt         string
	execution  string
	_eventId   string
	submit     string
	JSESSIONID string
}

type Grade struct {
	Items []*GradeItem `json:"items" binding:"required"`
}

type GradeItem struct {
	Kcmc   string `json:"kcmc" binding:"required"`
	Kcxzmc string `json:"kcxzmc" binding:"required"`
	Cj     string `json:"cj" binding:"required"`
	Jsxm   string `json:"jsxm" binding:"required"`
	Kclbmc string `json:"kclbmc" binding:"required"`
}

type Table struct {
	KbList []*TableItem `json:"kbList" binding:"required"`
}

type TableItem struct {
	kcmc string `json:"kcmc" binding:"required"`
	Zcd  string `json:"zcd" binding:"required"`
	Jcor string `json:"jcor" binding:"required"`
	Cdmc string `json:"cdmc" binding:"required"`
	// Kclbmc string `json:"kclbmc" binding:"required"`
}
