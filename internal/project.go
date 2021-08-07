package internal

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
}

func newProject(name, displayName string) Project {
	return Project{
		Name:        name,
		DisplayName: displayName,
	}
}
