Tax And Social Insurance Overdue Fine of PRC
中国税费款滞纳金计算程序
-----------------------------------------------

## config.xlsx 配置文件

其中的“每月缴费数据”表中填写当月的正税、社会保险费数据。

正税要压一个月写（因为下月缴上月税款），社会保险费写当月。

税款正常录入，社会保险费需填写**当年**社会保险费基数。

后面的缴费比例和单位缴费比例，税款全部填为 1。

## Build 编译方法

Windows 下面 `go build -ldflags='-H windowsgui' main.go`
