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

// GradeItem ... 成绩数据项，完整数据，带 id
type GradeItem struct {
	Kcmc   string `json:"kcmc" binding:"required"`   // 课程名称
	Kcxzmc string `json:"kcxzmc" binding:"required"` // 课程性质名称，比如专业主干课程/通识必修课
	Cj     string `json:"cj" binding:"required"`     // 成绩
	Xf     string `json:"xf" binding:"required"`     // 学分
	Jsxm   string `json:"jsxm" binding:"required"`   // 教师名称
	Kclbmc string `json:"kclbmc" binding:"required"` // 课程类别名称，比如专业课/公共课
	Kcgsmc string `json:"kcgsmc" binding:"required"` // 课程归属名称，比如文/理
	Kkbmmc string `json:"kkbmmc" binding:"required"` // 开课部门名称
	KchID  string `json:"kch_id" binding:"required"` // 课程号id
	JxbID  string `json:"jxb_id" binding:"required"` // 教学班id
	Zymc   string `json:"zymc" binding:"required"`   // 学生专业名称
	Jgmc   string `json:"jgmc" binding:"required"`   // 学生学院名称
}

type Table struct {
	KbList []*TableItem `json:"kbList" binding:"required"`
}

// TableItem ... 课表数据项，完整数据，带 id
type TableItem struct {
	Kcmc  string `json:"kcmc" binding:"required"`   // 课程名称
	Zcd   string `json:"zcd" binding:"required"`    // 课程教学周次，如：2-17周
	Xm    string `json:"xm" binding:"required"`     // 教师名
	Jcor  string `json:"jcor" binding:"required"`   // 课程教学节次，如：1-2
	Cdmc  string `json:"cdmc" binding:"required"`   // 教学地点
	Xqj   string `json:"xqj" binding:"required"`    // 星期几上课
	KchID string `json:"kch_id" binding:"required"` // 课程号id
	JxbID string `json:"jxb_id" binding:"required"` // 教学班id
	JghID string `json:"jgh_id" binding:"required"` // 不知道是啥id
}
