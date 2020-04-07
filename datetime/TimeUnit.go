package datetime

import (
	"fmt"
	"github.com/dlclark/regexp2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeUnit struct {
	TimeExpression string
	TimeNorm string
	timeFull []int
	timeOrigin []int
	time time.Time
	isAllDayTime bool
	isFirstTimeSolveContext bool
	noYear bool
	normalizer *TimeNormalizer

	_tp *TimePoint
	_tp_origin *TimePoint

	isMorning bool
	/*

	*/
}

func NewTimeUnitContext(expTime string, n *TimeNormalizer) (timeUnit *TimeUnit) {
	timeUnit = &TimeUnit{
		noYear: false,
		isMorning: false,
		isAllDayTime: true,
		isFirstTimeSolveContext: true,
		normalizer: n,
		TimeExpression: expTime,
	}
	return
}

func NewTimeUnit(expTime string, n *TimeNormalizer, contextTp *TimePoint) (timeUnit *TimeUnit){
	timeUnit = &TimeUnit{
		isAllDayTime: true,
		isFirstTimeSolveContext: true,
		_tp: NewTimePoint(),
		_tp_origin: contextTp,
		normalizer: n,
		TimeExpression: expTime,
	}
	timeUnit.TimeNormalization()
	return
}

func (tu *TimeUnit) TimeNormalization() {
	tu.normSetyear()
	tu.normSetMonth()
	tu.normSetDay()
	tu.normSetHour()
	copy(tu._tp_origin.tunit, tu._tp.tunit)

	//判断是时间点还是时间区间
	flag := true
	for i := 0; i < 5; i++ {
		if tu._tp.tunit[i] != -1 {
			flag = false
		}
	}
	if flag {
		tu.normalizer.isTimeSpan = true
	}
	if tu.normalizer.isTimeSpan {
		days := 0
		if tu._tp.tunit[0] > 0 {
			days += 365 *tu._tp.tunit[0]
		}
		if tu._tp.tunit[1] > 0 {
			days += 30 * tu._tp.tunit[1]
		}
		if tu._tp.tunit[2] > 0 {
			days += tu._tp.tunit[2]
		}
		tunit := tu._tp.tunit
		for i := 3; i < 6; i++ {
			if tu._tp.tunit[i] < 0 {
				tunit[i] = 0
			}
		}
		seconds := tunit[3] * 3600 + tunit[4] * 60 + tunit[5]
		if seconds == 0 && days == 0 {
			tu.normalizer.invalidSpan = true
		}
		tu.normalizer.timeSpan = tu.genSpan(days, seconds)
		return
	}

	time_grid := strings.Split(tu.normalizer.timeBase, "-")
	tunitpointer := 5
	for {
		if tunitpointer >= 0 && tu._tp.tunit[tunitpointer] < 0 {
			tunitpointer -= 1
		}
		break
	}
	for i := 0; i < tunitpointer; i++ {
		if tu._tp.tunit[i] < 0 {
			tg, err := strconv.Atoi(time_grid[i])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			tu._tp.tunit[i] = tg
		}
	}
	tu.time = tu.genTime(tu._tp.tunit)

}

func (tu *TimeUnit) genSpan(days, seconds int) string {
	day := (seconds / (3600 *24))
	h := (seconds %(3600 *24)) / 3600
	m := ((seconds % (3600 * 24)) % 3600) / 60
	s := ((seconds % (3600 * 24)) % 3600) % 60
	return   strconv.Itoa(days + day) + " days, " + fmt.Sprintf("%d:%02d:%02d", h, m , s)
}

func (tu *TimeUnit) genTime(tunit []int) time.Time{
	//t, _ := time.Parse("2016-01-02 15:04:05", "1970-01-01 00:00:00")
	dateValue := ""
	defaultDateValue := []string{"1970", "-01", "-01", " 00", ":00", ":00"}
	fmt.Println("tunit:", tunit)
	for i := 0; i < len(tunit); i++ {
		if tunit[i] > 0 {

			dateStr := strconv.Itoa(tunit[i])
			if tunit[i] < 10 {
				dateStr = "0" + dateStr
			}
			if i == 1 || i == 2 {
				dateValue += "-" + dateStr
			}
			if i == 4 || i == 5 {
				dateValue += "-" + dateStr
			}
			if i == 3 {
				dateValue += " " + dateStr
			}
			if i == 0 {
				dateValue += dateStr
			}
		} else {
			dateValue += defaultDateValue[i]
		}
	}
	fmt.Println("gentime datevalue:", dateValue)
	tn, err := time.Parse("2006-01-02 15:04:05", dateValue)
	if err != nil {
		fmt.Println("gen time error:", err.Error())
	}
	return tn
	/*if tunit[0] > 0 {

		dateValue += strconv.Itoa(tunit[0])
	} else {
		dateValue += "1970"
	}
	dateValue += "-"
	if tunit[1] > 0 {
		dateValue += strconv.Itoa(tunit[1])
	} else {
		dateValue += "01"
	}
	dateValue += "-"
	if tunit[2] > 0 {
		dateValue += strconv.Itoa(tunit[2])
	} else {
		dateValue += "01"
	}
	if tunit[3] > 0 {
		dateValue += strconv.Itoa(tunit[3])
	} else {
		dateValue += "00"
	}
	dateValue += ":"
	if tunit[4] > 0 {
		dateValue +=
	}*/

}

func (tu *TimeUnit) preferFuture(checkTimeIndex int) {
	/*
	如果用户选项是倾向于未来时间，检查checkTimeIndex所指的时间是否是过去的时间，如果是的话，将大一级的时间设为当前时间的+1。
        如在晚上说“早上8点看书”，则识别为明天早上;
        12月31日说“3号买菜”，则识别为明年1月的3号。
        :param checkTimeIndex: _tp.tunit时间数组的下标
        :return:
	*/
	/// # 1. 检查被检查的时间级别之前，是否没有更高级的已经确定的时间，如果有，则不进行处理.
	for i := 0; i < checkTimeIndex; i++ {
		if tu._tp.tunit[i] != -1 {
			return
		}
	}

	//根据上下文补充时间
	tu.checkContextTime(checkTimeIndex)

	/*
	3. 根据上下文补充时间后再次检查被检查的时间级别之前，是否没有更高级的已经确定的时间，如果有，则不进行倾向处理.
	*/

	//4.确认用户选项
	if !tu.normalizer.isPreferFuture {
		return
	}
	timeArr := strings.Split(tu.normalizer.timeBase, "-")
	cur, _ := time.Parse(tu.normalizer.timeBase, "2006-01-02-15-04-05")
	cu, _ := strconv.Atoi(timeArr[checkTimeIndex])

	if tu._tp.tunit[0] == -1 {
		tu.noYear = true
	} else {
		tu.noYear = false
	}

	if cu < tu._tp.tunit[checkTimeIndex] {
		return
	}

	//准备增加的时间单位是被检查的时间的上一级，将上一级时间+1
	cur = tu.addTime(cur, checkTimeIndex)
	timeArr = strings.Split(cur.Format("2006-01-02-15-04-05"), "-")
	for i := 0; i < checkTimeIndex; i++ {
		ta, _ := strconv.Atoi(timeArr[i])

		tu._tp.tunit[i] = ta
	}
}

func (tu *TimeUnit) addTime(cur time.Time, checkTimeIndex int) (cura time.Time) {
	if checkTimeIndex == 0 {
		cura = cur.AddDate(1, 0, 0)
	} else if checkTimeIndex == 1 {
		cura = cur.AddDate(0, 1, 0)
	} else if checkTimeIndex == 2 {
		cura = cur.AddDate(0, 0, 1)
	} else if checkTimeIndex == 3 {
		cura = cur.Add(time.Hour * 1)
	} else if checkTimeIndex == 4 {
		cura = cur.Add(time.Minute * 1)
	} else if checkTimeIndex == 5 {
		cura = cur.Add(time.Second * 1)
	} else {
		cura = cur
	}
	return cura
}

func (tu *TimeUnit) checkContextTime(checkTimeIndex int) {
	/*
	根据上下文时间补充时间信息
	*/
	if !tu.isFirstTimeSolveContext {
		return
	}
	for i := 0; i < checkTimeIndex; i++ {
		if tu._tp.tunit[i] == -1 && tu._tp_origin.tunit[i] != -1 {
			tu._tp.tunit[i] = tu._tp_origin.tunit[i]
		}
	}
	t_o := tu._tp_origin.tunit[checkTimeIndex]
	t := tu._tp.tunit[checkTimeIndex]
	if !tu.isMorning && checkTimeIndex == 3 && t_o >= 12 && (t_o-12) < t && t < 12 {
		tu._tp.tunit[checkTimeIndex] += 12
	}
	tu.isFirstTimeSolveContext = false
}

func (tu *TimeUnit) checkTime(parse []int) {
	timeArr := strings.Split(tu.normalizer.timeBase, "-")
	if tu.noYear {
		t, err := strconv.Atoi(timeArr[1])
		if err != nil {
			fmt.Println("checkTime err:", err.Error())
		}
		t2, err := strconv.Atoi(timeArr[2])
		if err != nil {
			fmt.Println("checkTime err:", err.Error())
		}
		if parse[1] == t {
			if parse[2] > t2 {
				parse[0] = parse[0] - 1
			}
 		}
		tu.noYear = false
	}
}

func (tu *TimeUnit) normCheckKeyWord() {
	rule := "凌晨"
	re := regexp.MustCompile(rule)
	match := re.FindString(tu.TimeExpression)

	if len(match) > 0 {
		tu.isMorning =  true
		if tu._tp.tunit[3] == -1 {
			tu._tp.tunit[3] = EarlyMoring
		} else if tu._tp.tunit[3] >= Noon && tu._tp.tunit[3] <= MidNigth {
			tu._tp.tunit[3] -= 12
		} else if tu._tp.tunit[3] == 0 {

		}
	}
}

func (tu *TimeUnit) normSetyear() {
	//只有两位数表示年份
	rule := "[0-9]{2}(年)"

	re := regexp.MustCompile(rule)

	match := re.FindString(tu.TimeExpression)
	fmt.Println("match:", match)
	match = strings.Replace(match, "年", "", -1)
	if len(match) > 0 {
		year, err := strconv.Atoi(match)
		if err != nil {
			fmt.Println("normSetYear1 Error:", err.Error())
			return
		}
		tu._tp.tunit[0] = year
		if tu._tp.tunit[0] >= 0 && tu._tp.tunit[0] < 100 {
			if tu._tp.tunit[0] < 30 {
				tu._tp.tunit[0] += 2000
			} else {
				tu._tp.tunit[0] += 1900
			}
		}
	}
	//三位数或四位数表示年份
	rule = "[0-9]?[0-9]{3}(年)"
	re = regexp.MustCompile(rule)
	match = re.FindString(tu.TimeExpression)
	match = strings.Replace(match, "年", "", -1)
	if len(match) > 0 {
		year, err := strconv.Atoi(match)
		if err != nil {
			fmt.Println("normSetYear2 Error:", err.Error())
			return
		}
		tu._tp.tunit[0] = year
	}
}

//月份处理
func (tu *TimeUnit) normSetMonth() {
	rule := "((10)|(11)|(12)|([1-9]))(月)"
	re := regexp.MustCompile(rule)
	match := re.FindString(tu.TimeExpression)
	match = strings.Replace(match, "月", "", -1)
	if len(match) > 0 {
		month, err := strconv.Atoi(match)
		if err != nil {
			fmt.Println("normSetMonth Error:", err.Error())
			return
		}
		tu._tp.tunit[1] = month
		tu.preferFuture(1)
	}

}

func (tu *TimeUnit) normSetDay() {
	/*
	日-规范化方法：该方法识别时间表达式单元的日字段
	*/
	rule := "([0-3][0-9]|[1-9])(日|号)"
	re := regexp.MustCompile(rule)
	match := re.FindString(tu.TimeExpression)
	match = strings.Replace(match, "日", "", -1)
	match = strings.Replace(match, "号", "", -1)
	if len(match) > 0 {
		day, err := strconv.Atoi(match)
		if err != nil {
			fmt.Println("normSetDay Error:", err.Error())
			return
		}
		tu._tp.tunit[2] = day
		tu.preferFuture(2)

	}
}

func (tu *TimeUnit) normSetHour() {
	rule := `(?<!(周|星期))([0-2]?[0-9])(?=(点|时))`
	//re := regexp.MustCompile(rule)
	re := regexp2.MustCompile(rule, regexp2.None)
	matchs, err := re.FindStringMatch(tu.TimeExpression)
	if err != nil {
		fmt.Println("normSetHour error:", err)
		return
	}
	if matchs == nil {
		fmt.Println("normSetHour matchs nil")
		return
	}
	tu.isAllDayTime = false
	match := matchs.Group.String()
	td, _ := strconv.Atoi(matchs.Group.String())
	//match := re.FindString(tu.TimeExpression)
	fmt.Println("normSethour match:", match, tu.TimeExpression)
	tu._tp.tunit[3] = td

}