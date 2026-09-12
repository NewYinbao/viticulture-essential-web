package game

// LibraryGroup contains only immutable card definitions, never room/deck state.
type LibraryGroup struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Cards []LibraryCard `json:"cards"`
}
type LibraryCard struct {
	Card
	RequiresTuscany bool `json:"requiresTuscany,omitempty"`
}

func CardLibrary() []LibraryGroup {
	groups := []LibraryGroup{}
	add := func(id, name string, cards []Card) {
		entries := make([]LibraryCard, 0, len(cards))
		for _, card := range cards {
			entries = append(entries, LibraryCard{Card: card})
		}
		groups = append(groups, LibraryGroup{ID: id, Name: name, Cards: entries})
	}
	add("ee", "Essential Edition", Catalog())
	add("moor", "Moor Visitors", MoorCatalog())
	add("rhine", "Rhine Valley", RhineCatalog(ExpansionConfig{Board: "tuscany"}))
	eeRhine := map[string]bool{}
	for _, card := range RhineCatalog(ExpansionConfig{Board: "ee"}) {
		eeRhine[card.ID] = true
	}
	for i := range groups[2].Cards {
		groups[2].Cards[i].RequiresTuscany = !eeRhine[groups[2].Cards[i].ID]
	}
	add("structures", "Tuscany 建筑", StructureCatalog())
	workers := []Card{}
	for _, worker := range SpecialWorkerCatalog() {
		workers = append(workers, Card{ID: "worker-" + worker.ID, Type: "worker", Name: worker.Name, Description: worker.Description, Implemented: true, RuleSource: worker.RuleSource})
	}
	add("workers", "特殊工人", workers)
	return groups
}
