package menu

type Menu struct {
	Id         int64   `json:"id"`
	Pid        int64   `json:"pid"`
	Title      string  `json:"title"`
	Icon       string  `json:"icon"`
	Url        string  `json:"url"`
	FilePath   string  `json:"file_ath"`
	Params     string  `json:"params"`
	Node       string  `json:"node"`
	Sort       int     `json:"sort"`
	Status     int     `json:"status"`
	CreateBy   int64   `json:"create_by"`
	IsInner    int     `json:"is_inner"`
	Values     string  `json:"values"`
	ShowSlider int     `json:"show_slider"`
	StatusText string  `json:"statusText"`
	InnerText  string  `json:"innerText"`
	FullUrl    string  `json:"fullUrl"`
	Children   []*Menu `json:"children"`
}
