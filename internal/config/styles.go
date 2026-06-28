package config

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed builtin_skins/*.yaml
var builtinSkinFiles embed.FS

// Skin 定义全部视觉样式参数。
type Skin struct {
	Body   BodyStyle   `yaml:"body"`
	Frame  FrameStyle  `yaml:"frame"`
	Table  TableStyle  `yaml:"table"`
	Prompt PromptStyle `yaml:"prompt"`
	Status StatusStyle `yaml:"status"`
	Form   FormStyle   `yaml:"form"`
}

type BodyStyle struct {
	FgColor Color `yaml:"fgColor"`
	BgColor Color `yaml:"bgColor"`
}

type FrameStyle struct {
	Border BorderStyle `yaml:"border"`
	Title  TitleStyle  `yaml:"title"`
}

type BorderStyle struct {
	FgColor    Color `yaml:"fgColor"`
	FocusColor Color `yaml:"focusColor"`
}

type TitleStyle struct {
	FgColor Color `yaml:"fgColor"`
}

type TableStyle struct {
	FgColor       Color       `yaml:"fgColor"`
	BgColor       Color       `yaml:"bgColor"`
	Header        HeaderStyle `yaml:"header"`
	CursorFgColor Color       `yaml:"cursorFgColor"`
	CursorBgColor Color       `yaml:"cursorBgColor"`
}

type HeaderStyle struct {
	FgColor Color `yaml:"fgColor"`
	BgColor Color `yaml:"bgColor"`
}

type PromptStyle struct {
	FgColor     Color `yaml:"fgColor"`
	BgColor     Color `yaml:"bgColor"`
	FilterColor Color `yaml:"filterColor"`
}

type StatusStyle struct {
	FgColor  Color `yaml:"fgColor"`
	BgColor  Color `yaml:"bgColor"`
	OkColor  Color `yaml:"okColor"`
	ErrColor Color `yaml:"errColor"`
}

type FormStyle struct {
	FgColor       Color `yaml:"fgColor"`
	BgColor       Color `yaml:"bgColor"`
	FieldFgColor  Color `yaml:"fieldFgColor"`
	FieldBgColor  Color `yaml:"fieldBgColor"`
	ButtonFgColor Color `yaml:"buttonFgColor"`
	ButtonBgColor Color `yaml:"buttonBgColor"`
}

// Styles 管理当前激活的皮肤。
type Styles struct {
	skin      *Skin
	themeDirs []string
}

// NewStyles 返回以内置默认皮肤初始化的 Styles。
func NewStyles() *Styles {
	return NewStylesWithOptions(Options{})
}

// NewStylesWithOptions 返回以运行时主题目录初始化的 Styles。
func NewStylesWithOptions(opts Options) *Styles {
	return &Styles{
		skin:      defaultSkin(),
		themeDirs: themeDirs(opts),
	}
}

// Load 从 skins 目录读取指定名称的皮肤。
// 传入 "" 或 "default" 可恢复内置皮肤。
func (s *Styles) Load(name string) error {
	if name == "" || name == "default" {
		s.skin = defaultSkin()
		return nil
	}
	if name == "teal" {
		s.skin = tealSkin()
		return nil
	}
	if data, ok := readBuiltinSkin(name); ok {
		skin := defaultSkin()
		if err := yaml.Unmarshal(data, skin); err != nil {
			return err
		}
		s.skin = skin
		return nil
	}
	for _, dir := range s.themeDirs {
		path := filepath.Join(dir, name+".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		skin := defaultSkin()
		if err := yaml.Unmarshal(data, skin); err != nil {
			return err
		}
		s.skin = skin
		return nil
	}
	return os.ErrNotExist
}

func (s *Styles) Body() BodyStyle     { return s.skin.Body }
func (s *Styles) Frame() FrameStyle   { return s.skin.Frame }
func (s *Styles) Table() TableStyle   { return s.skin.Table }
func (s *Styles) Prompt() PromptStyle { return s.skin.Prompt }
func (s *Styles) Status() StatusStyle { return s.skin.Status }
func (s *Styles) Form() FormStyle     { return s.skin.Form }

// AvailableSkins 返回 "default" 及 skins 目录下所有 *.yaml 名称。
func AvailableSkins() []string {
	return AvailableSkinsWithOptions(Options{})
}

// AvailableSkinsWithOptions 返回内置皮肤及默认/自定义皮肤目录中的皮肤。
func AvailableSkinsWithOptions(opts Options) []string {
	skins := []string{"default"}
	for _, name := range builtinSkinNames() {
		if !containsSkin(skins, name) {
			skins = append(skins, name)
		}
	}
	for _, dir := range themeDirs(opts) {
		appendSkinDir(&skins, dir)
	}
	return skins
}

func containsSkin(skins []string, name string) bool {
	for _, skin := range skins {
		if skin == name {
			return true
		}
	}
	return false
}

func themeDirs(opts Options) []string {
	dirs := []string{filepath.Join(cleanOrDefault(opts.ConfigDir, defaultConfigDir()), "skins")}
	for _, dir := range opts.ThemeDirs {
		if strings.TrimSpace(dir) != "" && !containsSkin(dirs, filepath.Clean(dir)) {
			dirs = append(dirs, filepath.Clean(dir))
		}
	}
	return dirs
}

func appendSkinDir(skins *[]string, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			name := strings.TrimSuffix(e.Name(), ".yaml")
			if !containsSkin(*skins, name) {
				*skins = append(*skins, name)
			}
		}
	}
}

func builtinSkinNames() []string {
	entries, err := fs.ReadDir(builtinSkinFiles, "builtin_skins")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return names
}

func readBuiltinSkin(name string) ([]byte, bool) {
	data, err := builtinSkinFiles.ReadFile(filepath.ToSlash(filepath.Join("builtin_skins", name+".yaml")))
	return data, err == nil
}

func defaultSkin() *Skin {
	return &Skin{
		Body: BodyStyle{FgColor: "#d0d0d0", BgColor: "#0d0d1a"},
		Frame: FrameStyle{
			Border: BorderStyle{FgColor: "#3a3a5c", FocusColor: "#6060c0"},
			Title:  TitleStyle{FgColor: "#8080ff"},
		},
		Table: TableStyle{
			FgColor:       "#d0d0d0",
			BgColor:       "#0d0d1a",
			CursorFgColor: "#0d0d1a",
			CursorBgColor: "#6060c0",
			Header:        HeaderStyle{FgColor: "#40a0ff", BgColor: "#0d0d1a"},
		},
		Prompt: PromptStyle{
			FgColor:     "#d0d0d0",
			BgColor:     "#1a1a30",
			FilterColor: "#ffff55",
		},
		Status: StatusStyle{
			FgColor:  "#606080",
			BgColor:  "#0d0d1a",
			OkColor:  "#40d040",
			ErrColor: "#d04040",
		},
		Form: FormStyle{
			FgColor:       "#d0d0d0",
			BgColor:       "#0d0d1a",
			FieldFgColor:  "#ffffff",
			FieldBgColor:  "#0d0d1a",
			ButtonFgColor: "#0d0d1a",
			ButtonBgColor: "#6060c0",
		},
	}
}

func tealSkin() *Skin {
	return &Skin{
		Body: BodyStyle{FgColor: "#b8d8d9", BgColor: "#00464d"},
		Frame: FrameStyle{
			Border: BorderStyle{FgColor: "#4fd9e6", FocusColor: "#66e7f0"},
			Title:  TitleStyle{FgColor: "#b980ff"},
		},
		Table: TableStyle{
			FgColor:       "#4fd9e6",
			BgColor:       "#00464d",
			CursorFgColor: "#00363b",
			CursorBgColor: "#55d7df",
			Header:        HeaderStyle{FgColor: "#66e7f0", BgColor: "#003b42"},
		},
		Prompt: PromptStyle{
			FgColor:     "#d8f3f4",
			BgColor:     "#003b42",
			FilterColor: "#ffc04d",
		},
		Status: StatusStyle{
			FgColor:  "#76aeb5",
			BgColor:  "#00464d",
			OkColor:  "#7ee787",
			ErrColor: "#ff5f5f",
		},
		Form: FormStyle{
			FgColor:       "#d8f3f4",
			BgColor:       "#00464d",
			FieldFgColor:  "#ecfeff",
			FieldBgColor:  "#00464d",
			ButtonFgColor: "#00363b",
			ButtonBgColor: "#ffc04d",
		},
	}
}
