package domain

type Analytics struct {
	Sum       float64
	Avg       float64
	Count     int64
	Median    float64
	Percent90 float64
	Details   []*Item
}
