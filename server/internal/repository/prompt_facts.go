package repository

import (
	"fmt"
	"time"
)

var weekdayZh = []string{"日", "一", "二", "三", "四", "五", "六"}

// appendPromptFactualContext 在系统提示末尾注入当前日期/时间等可核实事实，避免模型凭训练记忆编造「今天几号」等。
func appendPromptFactualContext(prompt string) string {
	now := time.Now().UTC()
	wd := int(now.Weekday())
	if wd < 0 || wd >= len(weekdayZh) {
		wd = 0
	}
	facts := fmt.Sprintf(`

【当前事实（用户问今天几号、星期几、现在几点等时必须严格以本节为准，勿凭训练记忆猜测；天气、新闻、股价等若本节未提供则说不知道，勿编造）】
今天是 %d年%d月%d日，星期%s。Today is %s. Current time (UTC): %s.`,
		now.Year(), int(now.Month()), now.Day(), weekdayZh[wd],
		now.Format("Monday, 2 January 2006"),
		now.Format("15:04"),
	)
	return prompt + facts
}
