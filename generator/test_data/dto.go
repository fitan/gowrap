package test_data

import gorm "gorm.io/gorm"

func queryDTO(v Query) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where("name = ?", v.Name)
		if v.Age != nil {
			db = db.Where("age > ?", v.Age)
		}
		db = db.Where("id IN ?", v.Ids)
		db = db.Where("model = ?", v.Model)
		db = db.Or(func(db *gorm.DB) *gorm.DB {
			db = db.Where("name = ?", v.Email.Name)
			db = db.Where("project_id = ?", func(db *gorm.DB) *gorm.DB {
				db = db.Where("id = ?", v.Email.Project.Id)
				db = db.Where("device_id = ?", func(db *gorm.DB) *gorm.DB {
					db = db.Where("port = ?", v.Email.Project.Device.Port)

					return db
				}(db.Session(&gorm.Session{NewDB: true}).Table("device").Select("id")))

				return db
			}(db.Session(&gorm.Session{NewDB: true}).Table("project").Select("id")))

			return db
		}(db.Session(&gorm.Session{NewDB: true})))
		db = db.Where("vm_uuid = ?", func(db *gorm.DB) *gorm.DB {
			db = db.Where("ip = ?", v.VM.Ip)

			return db
		}(db.Session(&gorm.Session{NewDB: true}).Table("vm").Select("uuid")))
		if v.PM != nil {
			db = db.Where(func(db *gorm.DB) *gorm.DB {
				if v.PM != nil {
					if v.NetMask != nil {
						db = db.Where("net_mask = ?", v.NetMask)
					}
					db = db.Where("limit = ?", v.Limit)

				}

				return db
			}(db.Session(&gorm.Session{NewDB: true})))

		}

		return db
	}
}
func copyDTO(src HelloRequestCopy) (dest HelloRequest) {
	dest = copyDTOObj{}.Copy(src)
	return
}

type copyDTOObj struct{}

func (d copyDTOObj) Copy(src HelloRequestCopy) (dest HelloRequest) {
	// basic =
	// basic = Name
	// Body.Name
	// src.Body.Name
	// dest.Body.Name
	dest.Body.Name = src.Body.Name
	// basic = Page
	// Paging.Page
	// src.Paging.Page
	// dest.Paging.Page
	dest.Paging.Page = src.Paging.Page
	// basic = Ip
	// Vm.Ip
	// src.Vm.Ip
	// dest.Vm.Ip
	dest.Vm.Ip = src.Vm.Ip
	/*
	   @dto-method fmt Sprintf
	*/
	// basic = HeaderName
	// HeaderName
	// src.HeaderName
	// dest.HeaderName
	dest.HeaderName = src.HeaderName
	/*
	   ID is the ID of the user.
	*/
	// basic = ID
	// ID
	// src.ID
	// dest.ID
	dest.ID = src.ID
	// basic = UUID
	// UUID
	// src.UUID
	// dest.UUID
	dest.UUID = src.UUID
	// basic = Time
	// Time
	// src.Time
	// dest.Time
	dest.Time = src.Time
	// basic = Age
	// Body.Age
	// src.Body.Age
	// dest.Body.Age
	dest.Body.Age = src.Body.Age
	// basic = Size
	// Paging.Size
	// src.Paging.Size
	// dest.Paging.Size
	dest.Paging.Size = src.Paging.Size
	// basic = Port
	// Vm.Port
	// src.Vm.Port
	// dest.Vm.Port
	dest.Vm.Port = src.Vm.Port
	// slice =
	dest.SayHi = src.SayHi
	/*
	   LastNames is the last names of the user.
	*/
	dest.LastNames = src.LastNames
	dest.LastNamesInt = src.LastNamesInt
	dest.Vm.NetWorks = src.Vm.NetWorks
	dest.Vm.VVMMSS = src.Vm.VVMMSS
	dest.VMS = src.VMS
	dest.Namespace = src.Namespace
	dest.ParentNames = src.ParentNames
	// map =
	dest.VMMap = src.VMMap
	// pointer =
	dest.ParentName = src.ParentName
	dest.FatherNames = src.FatherNames
	return
}
