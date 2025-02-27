package wpml

type ParentBuilder interface {
	IsTemplate() bool
	IsWayline() bool
}

type FolderBuilder struct {
	Folder *Folder
}

// AppendPlacemark 添加 Placemark
func (b *FolderBuilder) AppendPlacemark(p Placemark) *FolderBuilder {
	p.Index = len(b.Folder.Placemarks)
	b.Folder.Placemarks = append(b.Folder.Placemarks, p)
	return b
}

// AppendDefaultPlacemark 添加默认 Placemark
func (b *FolderBuilder) AppendDefaultPlacemark(lng, lat float64) *FolderBuilder {
	p := DefaultPlacemark(lng, lat)
	p.Index = len(b.Folder.Placemarks)
	b.Folder.Placemarks = append(b.Folder.Placemarks, p)
	return b
}
