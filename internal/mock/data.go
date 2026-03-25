package mock

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type MockData struct {
	Notes       []models.Note
	Accounts    []models.Profile
	Blocks      []models.Block
	BlockTypes  []models.BlockType
	BlockStates []models.BlockState
}

func NewMockData() *MockData {
	mock := &MockData{}
	mock.init()
	return mock
}

func basicState(block models.Block) models.BlockState {
	stateID := uuid.New()
	return models.BlockState{
		ID:      stateID,
		BlockID: block.ID,
		Formatting: map[string]interface{}{
			"format": "text",
			"tags":   []string{"пример", "тест", "мок"},
		},
		CreatedAt: time.Now().AddDate(0, -1, 0),
		UpdatedAt: time.Now(),
	}
}

func (m *MockData) init() {
	now := time.Now()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	m.Accounts = append(m.Accounts, models.Profile{
		ID:           userID,
		Username:     "testuser",
		Password:     []byte("hashed_password"),
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	m.BlockTypes = []models.BlockType{
		{ID: 1, Name: "text"},
		{ID: 2, Name: "image"},
		{ID: 3, Name: "code"},
		{ID: 4, Name: "quote"},
	}

	noteID := uuid.New()

	note := models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "The Forgotten City",
		ParentID:  nil,
		CreatedAt: now.AddDate(0, -1, 0),
		UpdatedAt: now,
	}
	m.Notes = append(m.Notes, note)

	blockID := uuid.New()
	block := models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "There is no game like Outer Wilds. That doesn’t stop fans from search for the elusive Wilds-like. One game that keeps popping up is The Forgotten City. Being very fond of flying into the sun and eating burned marshmallows, I was intrigued to try another knowledge based game.",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    1,
		Content:     "I’d like to start by stating my expectations when I opened TFC. Because I feel this review, and my experience with the game at large, are in big part a result of them. When the internets sold me on The Forgotten City, I was painted an image of a knowledge-based time loop mystery with lots of philosophy set in the roman empire. A period drama whodunnit Outer Wilds meets The Talos Principle?! Count me in! My experience was shot through this lens, with expectations you might’ve not had.",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "I usually begin reviews by talking about presentation. Alas, the graphics are often wonky and the mechanics are stiff. The NPC models are aggressively Bethesda-esque, and that’s not a compliment. Part of game’s fame comes from it starting life as a Skyrim mod made by 3 people, and unfortunately it shows. I can’t deny it’s an impressive feat. Few could create such an experience, but as I paid more for this than for some of my all-time favourite stories, I can’t see such context as an excuse for rocks with poor clipping and countless opportunities to get soft-locked in invisible walls. I think part of my negative perception in this regard stems from TFC going for a “realistic” look, which beside being subjectively boring is hard to do well on a budget. I genuinely think I would’ve liked the exact same game measurably more if it was styled as well as indies tend to be.",
		CreatedAt:   now.AddDate(0, -1, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))

	noteID = uuid.New()

	note = models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Our goals",
		ParentID:  nil,
		CreatedAt: now.AddDate(0, -2, 0),
		UpdatedAt: now,
	}
	m.Notes = append(m.Notes, note)

	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "People talk a lot these days about the need to have goals. Many books have been written about achieving goals. The speakers give a lot of advice on this topic. Do you have a goal?",
		CreatedAt:   now.AddDate(0, -2, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    1,
		Content:     "There are special technologies for achieving a goal. Successful people, when they try to achieve their goals, met many obstacles. In order not to stop halfway, they developed their own techniques. Their experience can serve as a good example for others. Let’s look at the important conditions for achieving the goal.",
		CreatedAt:   now.AddDate(0, -2, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "The goal should be written on paper. An unwritten goal is just a fantasy. When we write a goal on paper, we are sending a signal to our subconscious. From that moment on, the subconscious mind will be busy trying to find the best conditions for us to achieve a goal. This also keeps us motivated.",
		CreatedAt:   now.AddDate(0, -2, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))

	noteID = uuid.New()

	note = models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Eagle's history",
		ParentID:  nil,
		CreatedAt: now.AddDate(0, -3, 0),
		UpdatedAt: now,
	}
	m.Notes = append(m.Notes, note)

	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "The United States of America is a big country in Northern America. Today we will tell you about one of its most famous symbols — a bald eagle. Soon after the USA got its independence from Great Britain, the government decided to use its image on the Great Seal. The picture of a bald eagle is often used as a symbol of courage, strength and power.",
		CreatedAt:   now.AddDate(0, -3, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    1,
		Content:     "This bird lives only on the territory of Northern America, you won’t find it anywhere else. The eagle is very large: it may grow almost 3 feet high, its wingspan up to 8 feet. To tell the truth, the eagle isn’t really bald. Its head is covered with white feathers, that is why it seems to be bald.",
		CreatedAt:   now.AddDate(0, -3, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "The bird is a very committed partner in marriage. They choose marriage partners for life, and they take care for their babies together. Males and females look alike, but females are usually larger. Eagles build large nests, and usually they do it together. One of the biggest ones was recorded in the Guinness Book of Records, because it weighed almost 2 tons. These birds are one of a kind, because they can see even with their eyes closed. The thing is, in addition to usual eyelids they have special membranes on their eyes. Those membranes help them better preserve their eyes from the dust.",
		CreatedAt:   now.AddDate(0, -3, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))

	noteID = uuid.New()

	note = models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Metro and all about it",
		ParentID:  nil,
		CreatedAt: now.AddDate(0, -4, 0),
		UpdatedAt: now,
	}
	m.Notes = append(m.Notes, note)

	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Around the world, a lot of big cities have a metro. This mean of transport can carry many people and there are no traffic jams with this vehicle. The first metro was put into service in London in 1863. During the World War II London Underground was a shelter for people.",
		CreatedAt:   now.AddDate(0, -4, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    1,
		Content:     "When the wartime had been over, the metro began to appear in many cities. The cities grew, and the metro gave a solution to connect the centre of the city and its suburbs. Moreover, people could buy a ticket which had an affordable price. The metro could conquer the world fast, because it has a lot of advantages. It reduces road traffic, and it means that pollution is decreasing too.",
		CreatedAt:   now.AddDate(0, -4, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "Nowadays, the underground has more comfortable seats. There are escalators that can facilitate access to the platforms. Some stations have beautiful works of art. Their artists are not well-known as a rule, that is why these pictures have no great value. It helps to get rid of stealing. But the main thing is that they provide nice atmosphere. And it is very pleasant to be there. The Stockholm subway in Sweden is considered as the longest museum in the world.",
		CreatedAt:   now.AddDate(0, -4, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))

	noteID = uuid.New()

	note = models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Shopping",
		ParentID:  nil,
		CreatedAt: now.AddDate(0, -5, 0),
		UpdatedAt: now,
	}
	m.Notes = append(m.Notes, note)

	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "If we need to buy something, first of all we go to the shop. There are many different shops where you can buy whatever you want - from food to screws, bolts and nuts. It is not difficult to guess what type of store is the most popular. It may be said without exaggeration that these types of shops are supermarkets and grocery stores. A human being eats every day, so passing by such shops is a rather difficult thing.",
		CreatedAt:   now.AddDate(0, -5, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    1,
		Content:     "In every city you will find such shops as grocery stores, clothing stores, bakeries, butcheries. I love going to the flower shop most of all because flowers are my passion. Every week I go to an antique (curiosity) shop, because I really enjoy the original, ancient things. From time to time I visit the toy store in order to buy toys for my nephews and children. Almost every month I go to the gift shop so that I can buy gifts on birthday for my family and friends.",
		CreatedAt:   now.AddDate(0, -5, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
	blockID = uuid.New()
	block = models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "I like to spend my time on shopping, preferably I like the self-service shops. You can scrutinize something as long as you like. A nagging seller does not hurry you, you are your own master. After it all, you can calmly go to the cashier, where all purchases will be counted and added up. In our time, it’s not only supermarkets that work in such a way, but also department stores, clothing shops and household goods shops.",
		CreatedAt:   now.AddDate(0, -5, 0),
		UpdatedAt:   now,
	}
	m.Blocks = append(m.Blocks, block)
	m.BlockStates = append(m.BlockStates, basicState(block))
}

func (m *MockData) GetBlocksByNoteID(noteID uuid.UUID) []models.Block {
	var blocks []models.Block
	for _, block := range m.Blocks {
		if block.NoteID == noteID {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (m *MockData) GetBlockStatesByBlockID(blockID uuid.UUID) []models.BlockState {
	var states []models.BlockState
	for _, state := range m.BlockStates {
		if state.BlockID == blockID {
			states = append(states, state)
		}
	}
	return states
}
