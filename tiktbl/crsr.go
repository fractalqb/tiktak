package tiktbl

type Cursor struct {
	t     *Data
	r, c  int
	Error error
}

func (t *Data) At(r, c int) *Cursor { return &Cursor{t, r, c, nil} }

func (c *Cursor) Skip(i int) *Cursor {
	c.c += i
	if c.c < 0 {
		c.c = 0
	}
	return c
}

func (c *Cursor) SetString(text string, opts ...CellOpt) *Cursor {
	if c.Error != nil {
		return c
	}
	c.Error = c.t.SetString(c.r, c.c, text, opts...)
	if c.Error == nil {
		c.advance()
	}
	return c
}

func (c *Cursor) Set(data any, opts ...CellOpt) *Cursor {
	if c.Error != nil {
		return c
	}
	c.Error = c.t.Set(c.r, c.c, data, opts...)
	if c.Error == nil {
		c.advance()
	}
	return c
}

func (c *Cursor) advance() {
	cell := &c.t.rows[c.r][c.c]
	if cell.span > 0 {
		c.c += cell.span
	}
}

func (c *Cursor) NextRow() *Cursor {
	if c.Error != nil {
		return c
	}
	c.r, c.c = c.r+1, 0
	return c
}

func (c *Cursor) SetStrings(texts ...string) *Cursor {
	for _, txt := range texts {
		c = c.SetString(txt)
	}
	return c
}

func (c *Cursor) Sets(data ...any) *Cursor {
	for _, txt := range data {
		c = c.Set(txt)
	}
	return c
}

func (c *Cursor) With(opts ...CellOpt) with { return with{opts, c} }

type with struct {
	opts []CellOpt
	crsr *Cursor
}

func (w with) SetStrings(texts ...string) *Cursor {
	for _, text := range texts {
		w.crsr = w.crsr.SetString(text, w.opts...)
	}
	return w.crsr
}

func (w with) Sets(data ...any) *Cursor {
	for _, d := range data {
		w.crsr = w.crsr.Set(d, w.opts...)
	}
	return w.crsr
}
