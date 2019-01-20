package copier

import (
	refcopier "github.com/jinzhu/copier"
)

//Copy - Deep copies source to target struct.  Using github.com/jinszhu/copier for now, will allow replacement if needed + orders the parameters more intuitively
func Copy(src interface{}, target interface{}) {

	refcopier.Copy(target, src)

}
