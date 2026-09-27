package connect

import "testing"

import (
	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	Convey("Given some integer with a starting value", t, func() {
		url := "https://www.baidu.com"
		got := Get(url)

		Convey("When the integer is incremented", func() {

			Convey("The value should be greater by one", func() {
				// 断言
				So(got, ShouldEqual, true)
			})
		})
	})
}
