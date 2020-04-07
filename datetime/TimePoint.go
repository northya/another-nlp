package datetime

type TimePoint struct {
	tunit []int
}

func NewTimePoint() (timePoint *TimePoint){
	timePoint = &TimePoint{
		tunit: []int{-1, -1, -1, -1, -1, -1},
	}
	return
}