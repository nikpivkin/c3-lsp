package parser

import (
	"testing"

	"github.com/pherrymason/c3-lsp/lsp/document"
	"github.com/pherrymason/c3-lsp/lsp/symbols"
	idx "github.com/pherrymason/c3-lsp/lsp/symbols"
	"github.com/stretchr/testify/assert"
)

func TestParserModules_should_get_scopes_of_given_module(t *testing.T) {
	pm := NewParsedModules("a-doc")
	module := symbols.NewModuleBuilder("xxx", "a-doc").Build()
	pm.modules.Set("foo", &module)

	assert.Equal(t, &module, pm.Get("foo"))
}

func TestParserModules_GetLoadableModules_should_get_scopes_that_are_children_of_given_module(t *testing.T) {
	pm := NewParsedModules("a-doc")
	loadableModule := symbols.NewModuleBuilder("foo::bar", "a-doc").Build()
	pm.modules.Set("foo::bar", &loadableModule)
	loadableModule2 := symbols.NewModuleBuilder("foo", "a-doc").Build()
	pm.modules.Set("foo", &loadableModule2)
	notLoadableModule := symbols.NewModuleBuilder("yyy", "a-doc").Build()
	pm.modules.Set("yyy", &notLoadableModule)

	modules := pm.GetLoadableModules(symbols.NewModulePathFromString("foo"))

	assert.Equal(t, &loadableModule, modules[0])
	assert.Equal(t, &loadableModule2, modules[1])
	assert.Equal(t, 2, len(modules))
}

func TestParserModules_GetLoadableModules_should_get_scopes_that_are_parent_of_given_module(t *testing.T) {
	pm := NewParsedModules("a-doc")
	loadableModule := symbols.NewModuleBuilder("foo::bar", "a-doc").Build()
	pm.modules.Set("foo::bar", &loadableModule)
	loadableModule2 := symbols.NewModuleBuilder("foo", "a-doc").Build()
	pm.modules.Set("foo", &loadableModule2)
	notLoadableModule := symbols.NewModuleBuilder("yyy", "a-doc").Build()
	pm.modules.Set("yyy", &notLoadableModule)
	notLoadableModule2 := symbols.NewModuleBuilder("foo::circle", "a-doc").Build()
	pm.modules.Set("foo::circle", &notLoadableModule2)

	modules := pm.GetLoadableModules(symbols.NewModulePathFromString("foo::bar::line"))

	assert.Equal(t, &loadableModule, modules[0])
	assert.Equal(t, &loadableModule2, modules[1])
	assert.Equal(t, 2, len(modules))
}

func TestParserModules_HasImplicitLoadableModules_should_return_false_when_there_are_not_implicitly_loadable_modules(t *testing.T) {
	cases := []struct {
		desc            string
		searchingModule string
		existingModule  string
	}{
		{"no matching", "foo", "xxx"},
		{"cousing module", "foo::bar", "foo::circle"},
	}
	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			module := symbols.NewModulePathFromString(tt.searchingModule)
			pm := NewParsedModules("a-doc")
			loadableModule := symbols.NewModuleBuilder(tt.existingModule, "a-doc").Build()
			pm.modules.Set(tt.existingModule, &loadableModule)

			assert.False(t, pm.HasImplicitLoadableModules(module))
		})
	}
}

func TestParserModules_HasImplicitLoadableModules_should_return_true_when_there_is_same_module(t *testing.T) {
	module := symbols.NewModulePathFromString("foo")

	pm := NewParsedModules("a-doc")
	loadableModule := symbols.NewModuleBuilder("foo", "a-doc").Build()
	pm.modules.Set("foo", &loadableModule)

	assert.True(t, pm.HasImplicitLoadableModules(module))
}

func TestParserModules_HasImplicitLoadableModules_should_return_true_when_there_is_a_child_module(t *testing.T) {
	module := symbols.NewModulePathFromString("foo")

	pm := NewParsedModules("a-doc")
	loadableModule := symbols.NewModuleBuilder("foo::bar", "a-doc").Build()
	pm.modules.Set("foo::bar", &loadableModule)

	assert.True(t, pm.HasImplicitLoadableModules(module))
}

func TestParserModules_HasImplicitLoadableModules_should_return_true_when_there_is_a_parent_module(t *testing.T) {
	module := symbols.NewModulePathFromString("foo::bar")

	pm := NewParsedModules("a-doc")
	loadableModule := symbols.NewModuleBuilder("foo", "a-doc").Build()
	pm.modules.Set("foo", &loadableModule)

	assert.True(t, pm.HasImplicitLoadableModules(module))
}

func TestExtractSymbols_module_with_generics(t *testing.T) {

	//module std::atomic::types(<Type>);
	source := `module foo_test(<Type1, Type2>);
		struct Foo
		{
			Type1 a;
		}
		fn Type2 test(Type2 b, Foo *foo)
		{
			return foo.a + b;
		}
		
		module foo::another::deep(<Type>);
		int bar = 0;`

	doc := document.NewDocument("docid", source)
	parser := createParser()
	symbols := parser.ParseSymbols(&doc)

	module := symbols.Get("foo_test")
	assert.Equal(t, "foo_test", module.GetName())

	// Generic parameter was found
	generic, ok := module.GenericParameters["Type1"]
	assert.True(t, ok)
	assert.Equal(t, "Type1", generic.GetName())
	assert.Equal(t, idx.NewRange(0, 17, 0, 22), generic.GetIdRange())
	assert.Equal(t, idx.NewRange(0, 17, 0, 22), generic.GetDocumentRange())

	// Generic parameter was found
	generic, ok = module.GenericParameters["Type2"]
	assert.True(t, ok)
	assert.Equal(t, "Type2", generic.GetName())
	assert.Equal(t, idx.NewRange(0, 24, 0, 29), generic.GetIdRange())
	assert.Equal(t, idx.NewRange(0, 24, 0, 29), generic.GetDocumentRange())

	module = symbols.Get("foo::another::deep")
	assert.Equal(t, "foo::another::deep", module.GetName())
}
