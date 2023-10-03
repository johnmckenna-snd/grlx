package ingredients

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gogrlx/grlx/types"
)

var (
	ingTex sync.Mutex
	ingMap map[types.Ingredient]map[string]types.RecipeCooker
)

func init() {
	ingMap = make(map[types.Ingredient]map[string]types.RecipeCooker)
}

type MethodProps struct {
	Key   string
	Val   string
	IsReq bool
}

type MethodPropsSet []MethodProps

func (m MethodPropsSet) ToMap() map[string]string {
	ret := make(map[string]string)
	for _, v := range m {
		ret[v.Key] = v.Val
		if v.IsReq {
			ret[v.Key] = ret[v.Key] + ",req"
		} else {
			ret[v.Key] = ret[v.Key] + ",opt"
		}
	}
	return ret
}

func RegisterAllMethods(step types.RecipeCooker) {
	ingTex.Lock()
	defer ingTex.Unlock()
	name, methods := step.Methods()
	_, ok := ingMap[types.Ingredient(name)]
	if !ok {
		ingMap[types.Ingredient(name)] = make(map[string]types.RecipeCooker)
	}
	for _, method := range methods {
		ingMap[types.Ingredient(name)][method] = step
	}
}

var (
	ErrUnknownIngredient = errors.New("unknown ingredient")
	ErrUnknownMethod     = errors.New("unknown method")
)

func NewRecipeCooker(id types.StepID, ingredient types.Ingredient, method string, params map[string]interface{}) (types.RecipeCooker, error) {
	fmt.Printf("cooking %s %s %s\n", id, ingredient, method)
	ingTex.Lock()
	defer ingTex.Unlock()
	fmt.Printf("%v\n", ingMap)
	if r, ok := ingMap[ingredient]; ok {
		if ing, ok := r[method]; ok {
			return ing.Parse(string(id), method, params)
		}
		return nil, ErrUnknownMethod
	}
	return nil, ErrUnknownIngredient
}
