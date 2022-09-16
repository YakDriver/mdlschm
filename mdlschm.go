package mdlschm

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/numbervalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// Tag keys
	TagComputed            = "computed"
	TagOptional            = "optional"
	TagRequired            = "required"
	TagSensitive           = "sensitive"
	TagDeprecationMessage  = "deprecation"
	TagDescription         = "desc"
	TagMarkdownDescription = "md"
	TagPlanModifiers       = "pmods"
	TagSnakeName           = "snake"
	TagValidators          = "valid"
	TagVersion             = "version"
	TagCollection          = "collection"

	// Tag Values
	TagCollectionList = "list"
	TagCollectionSet  = "set"
	TagTrue           = "true"

	TagPlanModifierReplace = "replace"
	TagPlanModifierDefault = "default"
	TagPlanModifierUSFU    = "usfu"

	TagValidatorBetween = "between"
	TagValidatorOneOf   = "oneof"
	TagValidatorNoneOf  = "noneof"

	SpecialTypeBlock = "block"
)

type nest struct {
	schema     *tfsdk.Schema
	blocks     map[string]tfsdk.Block
	attributes map[string]tfsdk.Attribute
	block      *tfsdk.Block
	attribute  *tfsdk.Attribute
}

// New converts a model struct into a tfsdk.Schema using field types and tags
// as cues to the schema details. New supports arbitrary depth of nested
// structs. New also supports many but not all validators and plan modifiers.
func New(model any) tfsdk.Schema {
	if reflect.ValueOf(model).Kind() != reflect.Struct {
		panic(fmt.Sprintf("internal error (expected struct, got %s)", reflect.ValueOf(model).Kind()))
	}

	n := rAttribute(model, "", false, 0)

	if n.schema == nil {
		panic("no schema achieved")
	}

	e := reflect.ValueOf(model)

	for i := 0; i < e.NumField(); i++ {
		if !e.Type().Field(i).IsExported() && e.Type().Field(i).Name == "_" && e.Type().Field(i).Type.Kind() == reflect.Struct {
			// special field to define schema-level things, eg, markdown description
			schemaLevelOptions(n.schema, string(e.Type().Field(i).Tag))
			break
		}
	}

	//return tfsdk.Schema{
	//	Attributes: sAttributes(model),
	//}
	return *n.schema
}

// 				Nested Attributes	Nested Blocks
// Schema		Yes					Yes
// Attributes	Yes					No
// Blocks		Yes					Yes

func rAttribute(model any, tags string, fromSlice bool, level int) *nest {
	if l := leaf(model, tags); l != nil {
		n := nest{}
		addAttrOptions(l, tags, reflect.TypeOf(model).String())
		n.attribute = l
		return &n
	}

	switch reflect.ValueOf(model).Kind() {
	case reflect.Struct:
		attrs := make(map[string]tfsdk.Attribute)
		blocks := make(map[string]tfsdk.Block)

		e := reflect.ValueOf(model)

		for i := 0; i < e.NumField(); i++ {
			if !e.Type().Field(i).IsExported() {
				continue
			}

			s := snakeCase(e.Type().Field(i).Name, string(e.Type().Field(i).Tag))
			n := rAttribute(e.Field(i).Interface(), string(e.Type().Field(i).Tag), false, level+1)
			if n.attribute != nil {
				attrs[s] = *n.attribute
			}
			if n.block != nil {
				blocks[s] = *n.block
			}
		}

		if level == 0 {
			return schemaNest(&blocks, &attrs)
		} else {
			return blockNest(&blocks, &attrs, fromSlice, tags)
		}
	case reflect.Slice:
		if reflect.TypeOf(model).Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("unrecognized slice type: %s", reflect.TypeOf(model).Elem().Kind()))
		}

		return rAttribute(reflect.Zero(reflect.TypeOf(model).Elem()).Interface(), tags, true, level+1)
	case reflect.Map:
		panic("only maps with string keys are supported")
	default:
		e := reflect.ValueOf(model)
		panic(fmt.Sprintf("got unrecognized type: %v", e.Type()))
	}
}

func schemaNest(blocks *map[string]tfsdk.Block, attrs *map[string]tfsdk.Attribute) *nest {
	s := &tfsdk.Schema{}

	if len(*blocks) > 0 {
		s.Blocks = *blocks
	}

	if len(*attrs) > 0 {
		s.Attributes = *attrs
	}

	n := nest{}
	n.schema = s
	return &n
}

func blockNest(blocks *map[string]tfsdk.Block, attrs *map[string]tfsdk.Attribute, slice bool, tags string) *nest {
	b := &tfsdk.Block{}

	if len(*blocks) > 0 {
		b.Blocks = *blocks
	}

	if len(*attrs) > 0 {
		b.Attributes = *attrs
	}

	addBlockOptions(b, slice, tags)

	n := nest{}
	n.block = b
	return &n
}

func schemaLevelOptions(schm *tfsdk.Schema, tags string) {
	if v := tagValue(TagVersion, tags); v != "" {
		vi, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			panic(fmt.Sprintf("version must be an int, not %s: %s", v, err))
		}
		schm.Version = vi
	}

	if v := tagValue(TagDeprecationMessage, tags); v != "" {
		schm.DeprecationMessage = v
	}

	if v := tagValue(TagDescription, tags); v != "" {
		schm.Description = v
	}

	if v := tagValue(TagMarkdownDescription, tags); v != "" {
		schm.MarkdownDescription = v
	}
}

func leaf(model any, tags string) *tfsdk.Attribute {
	a := tfsdk.Attribute{}

	switch reflect.TypeOf(model).String() {
	case "types.Bool", "bool":
		a.Type = types.BoolType
		return &a
	case "types.Float64", "float", "float64":
		a.Type = types.Float64Type
		return &a
	case "types.Int64", "int64", "int":
		a.Type = types.Int64Type
		return &a
	case "types.Number":
		a.Type = types.NumberType
		return &a
	case "types.String", "string":
		a.Type = types.StringType
		return &a
	case "map[string]types.Bool", "map[string]bool":
		a.Type = types.MapType{
			ElemType: types.BoolType,
		}
		return &a
	case "map[string]types.Float64", "map[string]float", "map[string]float64":
		a.Type = types.MapType{
			ElemType: types.Float64Type,
		}
		return &a
	case "map[string]types.Int64", "map[string]int64", "map[string]int":
		a.Type = types.MapType{
			ElemType: types.Int64Type,
		}
		return &a
	case "map[string]types.Number":
		a.Type = types.MapType{
			ElemType: types.NumberType,
		}
		return &a
	case "map[string]types.String", "map[string]string":
		a.Type = types.MapType{
			ElemType: types.StringType,
		}
		return &a
	case "[]types.Bool", "[]bool":
		if tagValue(TagCollection, tags) == TagCollectionSet {
			a.Type = types.SetType{
				ElemType: types.BoolType,
			}
			return &a
		}

		a.Type = types.ListType{
			ElemType: types.BoolType,
		}
		return &a
	case "[]types.Float64", "[]float", "[]float64":
		if tagValue(TagCollection, tags) == TagCollectionSet {
			a.Type = types.SetType{
				ElemType: types.Float64Type,
			}
			return &a
		}

		a.Type = types.ListType{
			ElemType: types.Float64Type,
		}
		return &a
	case "[]types.Int64", "[]int64", "[]int":
		if tagValue(TagCollection, tags) == TagCollectionSet {
			a.Type = types.SetType{
				ElemType: types.Int64Type,
			}
			return &a
		}

		a.Type = types.ListType{
			ElemType: types.Int64Type,
		}
		return &a
	case "[]types.Number":
		if tagValue(TagCollection, tags) == TagCollectionSet {
			a.Type = types.SetType{
				ElemType: types.NumberType,
			}
			return &a
		}

		a.Type = types.ListType{
			ElemType: types.NumberType,
		}
		return &a
	case "[]types.String", "[]string":
		if tagValue(TagCollection, tags) == TagCollectionSet {
			a.Type = types.SetType{
				ElemType: types.StringType,
			}
			return &a
		}

		a.Type = types.ListType{
			ElemType: types.StringType,
		}
		return &a
	}

	return nil
}

func addAttrOptions(a *tfsdk.Attribute, tags, attrType string) {
	if tagValue(TagComputed, tags) == TagTrue {
		a.Computed = true
	}

	if tagValue(TagOptional, tags) == TagTrue {
		a.Optional = true
	}

	if tagValue(TagRequired, tags) == TagTrue {
		a.Required = true
	}

	if tagValue(TagComputed, tags) == "" && tagValue(TagOptional, tags) == "" && tagValue(TagRequired, tags) == "" {
		a.Optional = true // default if computed, nor optional, nor required is set
	}

	if tagValue(TagSensitive, tags) == TagTrue {
		a.Sensitive = true
	}

	if v := tagValue(TagDeprecationMessage, tags); v != "" {
		a.DeprecationMessage = v
	}

	if v := tagValue(TagDescription, tags); v != "" {
		a.Description = v
	}

	if v := tagValue(TagMarkdownDescription, tags); v != "" {
		a.MarkdownDescription = v
	}

	if v := tagValue(TagPlanModifiers, tags); v != "" {
		a.PlanModifiers = pMods(v, attrType)
	}

	if v := tagValue(TagValidators, tags); v != "" {
		a.Validators = validators(v, attrType, false, tags)
	}
}

func addBlockOptions(b *tfsdk.Block, slice bool, tags string) {
	if tagValue(TagCollection, tags) == TagCollectionSet {
		b.NestingMode = tfsdk.BlockNestingModeSet
	} else {
		b.NestingMode = tfsdk.BlockNestingModeList
	}

	if v := tagValue(TagDeprecationMessage, tags); v != "" {
		b.DeprecationMessage = v
	}

	if v := tagValue(TagDescription, tags); v != "" {
		b.Description = v
	}

	if v := tagValue(TagMarkdownDescription, tags); v != "" {
		b.MarkdownDescription = v
	}

	if v := tagValue(TagPlanModifiers, tags); v != "" {
		b.PlanModifiers = pMods(v, SpecialTypeBlock)
	}

	// called no matter what since some are added even when not explicitly requested
	b.Validators = validators(tagValue(TagValidators, tags), SpecialTypeBlock, slice, tags)
}

func pMods(tagValue, attrType string) []tfsdk.AttributePlanModifier {
	pm := []tfsdk.AttributePlanModifier{}

	if hasTagArg(TagPlanModifierReplace, tagValue) {
		pm = append(pm, resource.RequiresReplace())
	}

	if hasTagArg(TagPlanModifierUSFU, tagValue) {
		pm = append(pm, resource.UseStateForUnknown())
	}

	if hasTagArg(TagPlanModifierDefault, tagValue) {
		dv := tagArgs(TagPlanModifierDefault, tagValue)
		switch attrType {
		case "types.Bool", "bool":
			b, err := strconv.ParseBool(dv)
			if err != nil {
				panic(fmt.Sprintf("default value (%s) is not a bool: %s", dv, err))
			}

			pm = append(pm, DefaultValue(types.Bool{Value: b}))
		case "types.Float64", "float", "float64":
			f, err := strconv.ParseFloat(dv, 64)
			if err != nil {
				panic(fmt.Sprintf("default value (%s) is not a number: %s", dv, err))
			}

			pm = append(pm, DefaultValue(types.Float64{Value: f}))
		case "types.Int64", "int64":
			i, err := strconv.ParseInt(dv, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("default value (%s) is not a number: %s", dv, err))
			}

			pm = append(pm, DefaultValue(types.Int64{Value: i}))
		case "types.Number", "int":
			f, err := strconv.ParseFloat(dv, 64)
			if err != nil {
				panic(fmt.Sprintf("default value (%s) is not a number: %s", dv, err))
			}

			pm = append(pm, DefaultValue(types.Number{Value: big.NewFloat(f)}))
		case "types.String", "string":
			pm = append(pm, DefaultValue(types.String{Value: dv}))
		}
	}

	return pm
}

func validators(tagV, attrType string, fromSlice bool, tags string) []tfsdk.AttributeValidator {
	vals := []tfsdk.AttributeValidator{}

	if hasTagArg(TagValidatorBetween, tagV) {
		if v := betweenValidator(tagV, attrType, tags); v != nil {
			vals = append(vals, v)
		}
	}

	// magic defaults and shortcuts (required = size > 0, optional = size >= 0)
	if !hasTagArg(TagValidatorBetween, tagV) && attrType == SpecialTypeBlock { // between takes precedence
		// fromSlice	Set		Required	Optional
		// F			F						=> list.between(0,1)
		// F			F					T	=> list.between(0,1)
		// F			F		T				=> list.between(1,1)
		// F			F		T			T	=> list.between(1,1)	(wrong, req + opt, req takes precedence)
		// F			T						=> set.between(0,1)		(wrong, 1 len set makes no sense)
		// F			T					T	=> set.between(0,1)		(wrong, 1 len set makes no sense)
		// F			T		T				=> set.between(1,1)		(wrong, 1 len set makes no sense)
		// F			T		T			T	=> set.between(1,1)		(wrong, req + opt, 1 len set)
		// T			F						=> no validator, any size ok
		// T			F					T	=> no validator, any size ok
		// T			F		T				=> list.at_least(1)
		// T			F		T			T	=> list.at_least(1)		(wrong, req + opt, req takes precedence)
		// T			T						=> no validator, any size ok
		// T			T					T	=> no validator, any size ok
		// T			T		T				=> set.at_least(1)
		// T			T		T			T	=> set.at_least(1)		(wrong, req + opt, req takes precedence)

		min := 0
		if tagValue(TagRequired, tags) == TagTrue {
			min = 1
		}

		if tagValue(TagCollection, tags) == TagCollectionSet {
			if fromSlice && min > 0 {
				vals = append(vals, setvalidator.SizeAtLeast(min))
			}

			if !fromSlice {
				vals = append(vals, setvalidator.SizeBetween(min, 1))
			}
		}

		if tagValue(TagCollection, tags) != TagCollectionSet { // list is default
			if fromSlice && min > 0 {
				vals = append(vals, listvalidator.SizeAtLeast(min))
			}

			if !fromSlice {
				vals = append(vals, listvalidator.SizeBetween(min, 1))
			}
		}
	}

	if hasTagArg(TagValidatorOneOf, tagV) {
		if v := oneOfValidator(tagV, attrType, tags); v != nil {
			vals = append(vals, v)
		}
	}

	if hasTagArg(TagValidatorNoneOf, tagV) {
		if v := noneOfValidator(tagV, attrType, tags); v != nil {
			vals = append(vals, v)
		}
	}

	if len(vals) > 0 {
		return vals
	}
	return nil
}

func betweenValidator(betweenValue, attrType, tags string) tfsdk.AttributeValidator {
	ta := tagArgs(TagValidatorBetween, betweenValue)
	args := strings.Split(ta, ",")
	if len(args) != 2 {
		panic(fmt.Sprintf("%s requires 2 numeric args, got %d", TagValidatorBetween, len(args)))
	}

	nums := []float64{}
	for _, a := range args {
		n, err := strconv.ParseFloat(a, 64)
		if err != nil {
			panic(fmt.Sprintf("%s requires 2 numeric args: %s", TagValidatorBetween, err))
		}
		nums = append(nums, n)
	}

	switch attrType {
	case "[]types.Bool", "[]bool",
		"[]types.Float64", "[]float", "[]float64",
		"[]types.Int64", "[]int64", "[]int",
		"[]types.Number",
		"[]types.String", "[]string", SpecialTypeBlock:
		if tagValue(TagCollection, tags) == TagCollectionSet {
			return setvalidator.SizeBetween(int(nums[0]), int(nums[1]))
		}
		return listvalidator.SizeBetween(int(nums[0]), int(nums[1]))
	case "types.ListType":
		return listvalidator.SizeBetween(int(nums[0]), int(nums[1]))
	case "types.SetType":
		return setvalidator.SizeBetween(int(nums[0]), int(nums[1]))
	case "types.String", "string":
		return stringvalidator.LengthBetween(int(nums[0]), int(nums[1]))
	case "types.Float64", "float", "float64", "types.Number":
		return float64validator.Between(nums[0], nums[1])
	case "types.Int64", "int", "int64":
		return int64validator.Between(int64(nums[0]), int64(nums[1]))
	}

	return nil
}

func oneOfValidator(oneOfValue, attrType, tags string) tfsdk.AttributeValidator {
	ta := tagArgs(TagValidatorOneOf, oneOfValue)
	args := strings.Split(ta, ",")

	switch attrType {
	case "types.Float64", "float", "float64":
		nums := []float64{}
		for _, a := range args {
			n, err := strconv.ParseFloat(a, 64)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorOneOf, err))
			}
			nums = append(nums, n)
		}
		return float64validator.OneOf(nums...)
	case "types.Int64", "int64", "int":
		nums := []int64{}
		for _, a := range args {
			n, err := strconv.ParseInt(a, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorOneOf, err))
			}
			nums = append(nums, n)
		}
		return int64validator.OneOf(nums...)
	case "types.Number":
		nums := []*big.Float{}
		for _, a := range args {
			bf := big.NewFloat(0.0)
			bf, _, err := bf.Parse(a, 10)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorOneOf, err))
			}
			nums = append(nums, bf)
		}
		return numbervalidator.OneOf(nums...)
	case "types.String", "string":
		return stringvalidator.OneOf(args...)
	}

	return nil
}

func noneOfValidator(noneOfValue, attrType, tags string) tfsdk.AttributeValidator {
	ta := tagArgs(TagValidatorNoneOf, noneOfValue)
	args := strings.Split(ta, ",")

	switch attrType {
	case "types.Float64", "float", "float64":
		nums := []float64{}
		for _, a := range args {
			n, err := strconv.ParseFloat(a, 64)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorNoneOf, err))
			}
			nums = append(nums, n)
		}
		return float64validator.NoneOf(nums...)
	case "types.Int64", "int64", "int":
		nums := []int64{}
		for _, a := range args {
			n, err := strconv.ParseInt(a, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorNoneOf, err))
			}
			nums = append(nums, n)
		}
		return int64validator.NoneOf(nums...)
	case "types.Number":
		nums := []*big.Float{}
		for _, a := range args {
			bf := big.NewFloat(0.0)
			bf, _, err := bf.Parse(a, 10)
			if err != nil {
				panic(fmt.Sprintf("%s requires numeric args: %s", TagValidatorNoneOf, err))
			}
			nums = append(nums, bf)
		}
		return numbervalidator.NoneOf(nums...)
	case "types.String", "string":
		return stringvalidator.NoneOf(args...)
	}

	return nil
}

func tagValue(key string, allTags string) string {
	tags := splitTags(allTags)

	for _, tag := range tags {
		parts := strings.Split(tag, ":")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == key {
			return strings.TrimPrefix(strings.TrimSuffix(parts[1], "\""), "\"")
		}
	}

	return ""
}

func hasTagArg(needle, haystack string) bool {
	h := splitTagValues(haystack)

	for _, check := range h {
		if check == needle || strings.HasPrefix(check, fmt.Sprintf("%s(", needle)) {
			return true
		}
	}

	return false
}

func tagArgs(needle, haystack string) string {
	h := splitTagValues(haystack)

	for _, check := range h {
		if strings.HasPrefix(check, fmt.Sprintf("%s(", needle)) {
			return strings.TrimPrefix(strings.TrimSuffix(check, ")"), fmt.Sprintf("%s(", needle))
		}
	}

	return ""
}

func splitTagValues(s string) []string {
	re := regexp.MustCompile(`(\([^\)]*),([^\)]*\))`)

	// extra juggling due to go's lack of lookahead in regex
	result := re.ReplaceAllString(s, "$1|||||$2")

	for true {
		newResult := re.ReplaceAllString(result, "$1|||||$2")
		if newResult != result {
			result = newResult
		} else {
			break
		}
	}

	p := []string{}

	h := strings.Split(result, ",")
	for _, v := range h {
		p = append(p, strings.Replace(v, "|||||", ",", -1))
	}

	return p
}

func splitTags(s string) []string {
	re := regexp.MustCompile(`(:"[^"]*) ([^"]*")`)

	// extra juggling due to go's lack of lookahead in regex
	result := re.ReplaceAllString(s, "$1|||||$2")

	for true {
		newResult := re.ReplaceAllString(result, "$1|||||$2")
		if newResult != result {
			result = newResult
		} else {
			break
		}
	}

	p := []string{}

	h := strings.Split(result, " ")
	for _, v := range h {
		p = append(p, strings.Replace(v, "|||||", " ", -1))
	}

	return p
}

func snakeCase(camel string, allTags string) string {
	snakeName := tagValue(TagSnakeName, allTags)

	if snakeName != "" {
		return snakeName
	}

	//preclean
	camel = strings.Replace(camel, "IDs", "Ids", -1)

	re := regexp.MustCompile(`([a-z])([A-Z]{2,})`)
	camel = re.ReplaceAllString(camel, `${1}_${2}`)

	re2 := regexp.MustCompile(`([A-Z][a-z])`)

	return strings.TrimPrefix(strings.ToLower(re2.ReplaceAllString(camel, `_$1`)), "_")
}
