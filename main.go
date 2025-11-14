package main

import (
	"fmt"
	i18n "i18n-tools/core"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	i18n.MustLoadI18n()
	trans := i18n.Trans("en", "success")
	fmt.Println(trans)

}
