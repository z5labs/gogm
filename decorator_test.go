package gogm

type validStruct struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	IndexField string `gogm:"index;name=index_field"`
	UniqueField int `gogm:"unique;name=unique_field"`
	OneToOne interface{} `gogm:"relationship=one2one;direction=incoming"`
	ManyToOne []interface{} `gogm:"relationship=many2one;direction=outgoing"`
	Props map[string]string `gogm:"properties"`
	IgnoreMe int `gogm:"-"`
}

//issue is that it has no id defined
type mostlyValidStruct struct{
	IndexField string `gogm:"index;name=index_field"`
	UniqueField int `gogm:"unique;name=unique_field"`
}

//nothing defined
type emptyStruct struct {}

//has a valid field but also has a messed up one
type invalidStructDecorator struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	MessedUp int `gogm:"sdfasdfasdfa"`
}

type invalidStructProperties struct {
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	Props map[string]string `gogm:"name=props"` //should have properties decorator
}