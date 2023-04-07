package tiktbl

type Styler interface {
	CellOpt
	style(string) string
}

func NoStyle() Styler { return Styles{} }

func AddStyles(s Styler, add ...Styler) Styler {
	styles, ok := s.(Styles)
	if !ok {
		styles = Styles{s}
	}
	for _, a := range add {
		if s, ok := a.(Styles); ok {
			styles = append(styles, s...)
		} else {
			styles = append(styles, a)
		}
	}
	if len(styles) == 1 {
		return styles[0]
	}
	return styles
}

type Style func(string) string

func (s Style) style(txt string) string { return s(txt) }

func (s Style) applyToCell(c *cell) error {
	c.styler = s
	return nil
}

type Styles []Styler

func (s Styles) style(txt string) string {
	for _, e := range s {
		txt = e.style(txt)
	}
	return txt
}

func (s Styles) applyToCell(c *cell) error {
	c.styler = s
	return nil
}
