package apperix

type Permissions struct {
	Create bool
	Read bool
	Update bool
	Delete bool
	Patch bool
	ReadHeaders bool
	ReadOptions bool
	Purge bool
	Copy bool
	Move bool
	Link bool
	Unlink bool
	Lock bool
	Unlock bool
	ReadProperties bool
	UpdateProperties bool
	CreateCollection bool
}

func (perm *Permissions) AllowAll() {
	perm.Create = true
	perm.Read = true
	perm.Update = true
	perm.Delete = true
	perm.Patch = true
	perm.ReadHeaders = true
	perm.ReadOptions = true
	perm.Purge = true
	perm.Copy = true
	perm.Move = true
	perm.Link = true
	perm.Unlink = true
	perm.Lock = true
	perm.Unlock = true
	perm.ReadProperties = true
	perm.UpdateProperties = true
	perm.CreateCollection = true
}

func (perm *Permissions) ForbidAll() {
	perm.Create = false
	perm.Read = false
	perm.Update = false
	perm.Delete = false
	perm.Patch = false
	perm.ReadHeaders = false
	perm.ReadOptions = false
	perm.Purge = false
	perm.Copy = false
	perm.Move = false
	perm.Link = false
	perm.Unlink = false
	perm.Lock = false
	perm.Unlock = false
	perm.ReadProperties = false
	perm.UpdateProperties = false
	perm.CreateCollection = false
}

func (perm *Permissions) Serialize() (mask uint32) {
	if perm.Create {
		mask |= (1 << 0)
	}
	if perm.Read {
		mask |= (1 << 1)
	}
	if perm.Update {
		mask |= (1 << 2)
	}
	if perm.Delete {
		mask |= (1 << 3)
	}
	if perm.Patch {
		mask |= (1 << 4)
	}
	if perm.ReadHeaders {
		mask |= (1 << 5)
	}
	if perm.ReadOptions {
		mask |= (1 << 6)
	}
	if perm.Purge {
		mask |= (1 << 7)
	}
	if perm.Copy {
		mask |= (1 << 8)
	}
	if perm.Move {
		mask |= (1 << 9)
	}
	if perm.Link {
		mask |= (1 << 10)
	}
	if perm.Unlink {
		mask |= (1 << 11)
	}
	if perm.Lock {
		mask |= (1 << 12)
	}
	if perm.Unlock {
		mask |= (1 << 13)
	}
	if perm.ReadProperties {
		mask |= (1 << 14)
	}
	if perm.UpdateProperties {
		mask |= (1 << 15)
	}
	if perm.CreateCollection {
		mask |= (1 << 16)
	}
	return mask
}

func (perm *Permissions) Deserialize(mask uint32) {
	if mask & (1 << 0) > 0 {
		perm.Create = true
	}
	if mask & (1 << 1) > 0 {
		perm.Read = true
	}
	if mask & (1 << 2) > 0 {
		perm.Update = true
	}
	if mask & (1 << 3) > 0 {
		perm.Delete = true
	}
	if mask & (1 << 4) > 0 {
		perm.Patch = true
	}
	if mask & (1 << 5) > 0 {
		perm.ReadHeaders = true
	}
	if mask & (1 << 6) > 0 {
		perm.ReadOptions = true
	}
	if mask & (1 << 7) > 0 {
		perm.Purge = true
	}
	if mask & (1 << 8) > 0 {
		perm.Copy = true
	}
	if mask & (1 << 9) > 0 {
		perm.Move = true
	}
	if mask & (1 << 10) > 0 {
		perm.Link = true
	}
	if mask & (1 << 11) > 0 {
		perm.Unlink = true
	}
	if mask & (1 << 12) > 0 {
		perm.Lock = true
	}
	if mask & (1 << 13) > 0 {
		perm.Unlock = true
	}
	if mask & (1 << 14) > 0 {
		perm.ReadProperties = true
	}
	if mask & (1 << 15) > 0 {
		perm.UpdateProperties = true
	}
	if mask & (1 << 16) > 0 {
		perm.CreateCollection = true
	}
}

