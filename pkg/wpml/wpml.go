package wpml

type Builder struct {
	Template *TemplateBuilder
	Wayline  *WaylineBuilder
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Init(author string) *Builder {
	b.Template = NewTemplateBuilder(author)
	return b
}

func (b *Builder) SetMissionConfig(config MissionConfig) *Builder {
	b.Template.SetMissionConfig(config)
	return b
}

func (b *Builder) SetDefaultMissionConfig(drone DroneInfo, payload PayloadInfo) *Builder {
	b.Template.SetMissionConfig(DefaultMissionConfig(drone, payload))
	return b
}

func (b *Builder) GenerateWayline() *Builder {
	b.Wayline = FromTemplate(*b.Template)
	return b
}
