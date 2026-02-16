package domain

type ItemAnalytics struct {
	Sum       float64
	Avg       float64
	Count     int64
	Median    float64
	Percent90 float64
}

type Analytics struct {
	Income  *ItemAnalytics
	Expense *ItemAnalytics
	Details []*Item
}
