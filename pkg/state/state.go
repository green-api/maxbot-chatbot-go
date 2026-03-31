package state

type BotApp interface {
	Reply(chatID int64, text string) error
}

type Scene interface {
	Start(app BotApp)
}

type State interface {
	getData() map[string]any
	setData(data map[string]any)
	updateData(data map[string]any)
	getScene() Scene
	setScene(scene Scene)
}

type MapState struct {
	data  map[string]any
	scene Scene
}

func NewMapState(data map[string]any, scene Scene) *MapState {
	newData := make(map[string]any)
	for key, value := range data {
		newData[key] = value
	}
	return &MapState{
		data:  newData,
		scene: scene,
	}
}

func (s *MapState) getData() map[string]any {
	return s.data
}

func (s *MapState) setData(data map[string]any) {
	s.data = data
}

func (s *MapState) updateData(data map[string]any) {
	for key, value := range data {
		s.data[key] = value
	}
}

func (s *MapState) getScene() Scene {
	return s.scene
}

func (s *MapState) setScene(scene Scene) {
	s.scene = scene
}
