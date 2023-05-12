package forms

type SystemFileForm struct {
	Name        string `form:"name" binding:"max=100"`
	IsDirectory *bool  `form:"is_directory" binding:"required"`
}
