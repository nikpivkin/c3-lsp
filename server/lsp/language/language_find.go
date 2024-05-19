package language

import (
	"fmt"

	"github.com/pherrymason/c3-lsp/lsp/parser"
	"github.com/pherrymason/c3-lsp/lsp/search_params"
	"github.com/pherrymason/c3-lsp/lsp/symbols"
	"github.com/pherrymason/c3-lsp/option"
)

func (l *Language) findModuleInPosition(docId string, position symbols.Position) string {
	for id, modulesByDoc := range l.parsedModulesByDocument {
		if id == docId {
			continue
		}

		for _, scope := range modulesByDoc.Modules() {
			if scope.GetDocumentRange().HasPosition(position) {
				return scope.GetModule().GetName()
			}
		}
	}

	panic("Module not found in position")
}

func (l *Language) implicitImportedParsedModules(acceptedModulePaths []symbols.ModulePath, excludeDocId option.Option[string]) []*symbols.Module {
	//var collectionParsedModules []*parser.ParsedModules
	var collectionModules []*symbols.Module
	for docId, parsedModules := range l.parsedModulesByDocument {
		if excludeDocId.IsSome() && excludeDocId.Get() == docId {
			continue
		}

		for _, scope := range parsedModules.Modules() {
			for _, acceptedModule := range acceptedModulePaths {
				if scope.GetModule().IsImplicitlyImported(acceptedModule) {
					collectionModules = append(collectionModules, scope)
					break
				}
			}
		}
	}

	return collectionModules
}

// Finds the closest selectedSymbol based on current scope.
// If not present in current Scope:
// - Search in files of same module
// - SearchParams in imported files (TODO)
// - SearchParams in global symbols in workspace
func (l *Language) findClosestSymbolDeclaration(searchParams search_params.SearchParams, debugger FindDebugger) option.Option[symbols.Indexable] {
	if IsLanguageKeyword(searchParams.Symbol()) {
		l.debug("Ignore because C3 keyword", debugger)
		return option.None[symbols.Indexable]()
	}

	l.debug(fmt.Sprintf("findClosestSymbolDeclaration on doc %s: %s: %s", searchParams.DocId(), searchParams.Module(), searchParams.Symbol()), debugger)

	/*position := searchParams.symbolRange.Start*/
	// Check if there's parent contextual information in searchParams
	if searchParams.HasAccessPath() {
		identifier := l.findInParentSymbols(searchParams, debugger)
		if identifier.IsSome() {
			return identifier
		}
	}

	if searchParams.HasModuleSpecified() {
		/*symbol := l._findSymbolDeclarationInModule(searchParams, debugger.goIn())
		if symbol != nil {
			return symbol
		}

		return nil
		*/
	}

	docIdOption := searchParams.DocId()
	var collectionParsedModules []parser.ParsedModules
	if docIdOption.IsSome() {
		docId := docIdOption.Get()
		parsedModules, found := l.parsedModulesByDocument[docId]
		if !found {
			return option.None[symbols.Indexable]()
		}

		collectionParsedModules = append(collectionParsedModules, parsedModules)
	} else {
		// Doc id not specified, search by module. Collect scope belonging to same module as searchParams.module
		for docId, parsedModules := range l.parsedModulesByDocument {
			if searchParams.ShouldExcludeDocId(docId) {
				continue
			}

			for _, scope := range parsedModules.Modules() {
				if scope.GetModule().IsImplicitlyImported(searchParams.ModulePath()) {
					collectionParsedModules = append(collectionParsedModules, parsedModules)
					break
				}
			}
		}
	}

	trackedModules := searchParams.TrackTraversedModules()
	var imports []string
	importsAdded := make(map[string]bool)
	for _, parsedModules := range collectionParsedModules {
		for _, scopedTree := range parsedModules.GetLoadableModules(searchParams.ModulePath()) {
			l.debug(fmt.Sprintf("Checking module \"%s\"", scopedTree.GetModuleString()), debugger)
			// Go through every element defined in scopedTree
			identifier, _ := findDeepFirst(
				searchParams.Symbol(),
				searchParams.SymbolPosition(),
				scopedTree,
				0,
				searchParams.IsLimitSearchInScope(),
				searchParams.ScopeMode(),
			)

			if identifier != nil {
				return option.Some(identifier)
			}

			// Not found, store imports traversed to avoid checking them again
			for _, imp := range scopedTree.Imports {
				if !importsAdded[imp] {
					imports = append(imports, imp)
				}
			}
		}
	}

	if searchParams.ContinueOnModules() {
		sb := search_params.NewSearchParamsBuilder().
			WithSymbol(searchParams.Symbol()).
			WithSymbolModule(searchParams.ModulePath()).
			WithExcludedDocs(searchParams.DocId()).
			WithScopeMode(search_params.InModuleRoot)
		searchInSameModule := sb.Build()

		found := l.findClosestSymbolDeclaration(searchInSameModule, debugger.goIn())
		if found.IsSome() {
			return found
		}
	}

	// Try to find element in one of the imported modules
	if docIdOption.IsSome() && len(imports) > 0 {
		for i := 0; i < len(imports); i++ {
			if !searchParams.TrackTraversedModule(imports[i]) {
				continue
			}

			module := imports[i]
			sp := search_params.NewSearchParamsBuilder().
				WithSymbol(searchParams.Symbol()).
				WithSymbolModule(symbols.NewModulePathFromString(module)).
				WithTrackedModules(trackedModules).
				Build()

			l.debug(fmt.Sprintf("findClosestSymbolDeclaration: search in imported module \"%s\": %s", module, searchParams.Symbol()), debugger)
			symbol := l.findSymbolDeclarationInModule(sp, debugger.goIn())
			if symbol.IsSome() {
				return symbol
			}
		}
	}

	// Not found...
	return option.None[symbols.Indexable]()
}

// Search symbols inside a given module
func (l *Language) findSymbolDeclarationInModule(searchParams search_params.SearchParams, debugger FindDebugger) option.Option[symbols.Indexable] {
	//expectedModule := searchParams.ModulePath().GetName()

	for docId, modulesByDoc := range l.parsedModulesByDocument {
		for _, scope := range modulesByDoc.GetLoadableModules(searchParams.ModulePath()) {
			//if scope.GetModuleString() != expectedModule { // TODO Ignore current doc we are comming from
			//	continue
			//}

			if !searchParams.TrackTraversedModule(scope.GetModuleString()) {
				continue
			}
			l.debug(fmt.Sprintf("findSymbolDeclarationInModule: search symbols in module \"%s\" file \"%s\"", scope.GetModuleString(), docId), debugger)

			sp := search_params.NewSearchParamsBuilder().
				WithSymbol(searchParams.Symbol()).
				WithDocId(docId).
				WithTrackedModules(searchParams.TrackedModules()).
				Build()

			symbol := l.findClosestSymbolDeclaration(
				sp,
				/*SearchParams{
					selectedToken:     searchParams.selectedToken,
					docId:             docId,
					scopeMode:         searchParams.scopeMode,
					continueOnModules: true,
					trackedModules:    searchParams.trackedModules,
				}*/FindDebugger{depth: debugger.depth + 1})
			l.debug(fmt.Sprintf("end searching symbols in module \"%s\" file \"%s\"", scope.GetModuleString(), docId), debugger)
			if symbol.IsSome() {
				return symbol
			}
		}
	}

	return option.None[symbols.Indexable]()
}

func findDeepFirst(identifier string, position symbols.Position, node symbols.Indexable, depth uint, limitSearchInScope bool, scopeMode search_params.ScopeMode) (symbols.Indexable, uint) {
	/*if limitSearchInScope {
		_, ok := node.(*symbols.Function)
		if ok && !node.GetDocumentRange().HasPosition(position) {
			return nil, depth
		}
	}*/
	/*
		if limitSearchInScope &&
			!node.GetDocumentRange().HasPosition(position) {
			return nil, depth
		}
	*/

	//if node.GetDocumentRange().HasPosition(position) {
	// Iterate first children with more children
	if scopeMode != search_params.InModuleRoot {
		for _, child := range node.NestedScopes() {
			// Check the fn itself! Maybe we are searching for it!
			if child.GetName() == identifier {
				return child, depth
			}

			if limitSearchInScope &&
				!child.GetDocumentRange().HasPosition(position) {
				return nil, depth
			}

			if result, resultDepth := findDeepFirst(identifier, position, child, depth+1, limitSearchInScope, scopeMode); result != nil {
				return result, resultDepth
			}
		}
	}

	for _, child := range node.ChildrenWithoutScopes() {
		if result, resultDepth := findDeepFirst(identifier, position, child, depth+1, limitSearchInScope, scopeMode); result != nil {
			return result, resultDepth
		}
	}

	// All elements found in nestable symbols checked, check node itself
	if node.GetName() == identifier {
		_, ok := node.(*symbols.Module) // Modules will be searched later explicitly.
		if !ok {
			return node, depth
		}
	}

	return nil, depth
}
