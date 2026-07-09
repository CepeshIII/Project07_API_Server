package db

import (
	"context"
	"course/api_server/internal/store"
	"fmt"
	"math/rand"
)

var usernames = []string{
	"alice",
	"bob",
	"charlie",
	"david",
	"eve",
	"frank",
	"grace",
	"heidi",
	"ivan",
	"judy",
	"mallory",
	"oscar",
	"peggy",
	"trent",
	"victor",
	"walter",
	"zara",
	"vlad",
	"viktoria",
	"donalt",
	"joseph",
}

var postTitles = []string{
	"The Future of Space Travel",
	"Healthy Sleep Habits",
	"The History of Coffee",
	"Understanding Compound Interest",
	"Deep Sea Exploration",
	"Artificial Intelligence Evolution",
	"Quantum Computing Basics",
	"The Secrets of Ancient Egypt",
	"Renewable Energy Growth",
	"The Mechanics of Photosynthesis",
	"The Art of Origami",
	"Cryptocurrency and Blockchain",
	"Volcanic Eruptions Explained",
	"The Human Microbiome",
	"The Great Barrier Reef",
	"Renaissance Art Innovation",
	"Electric Vehicle Tech",
	"The Mystery of Black Holes",
	"Mindfulness and Meditation",
	"Urban Farming Solutions",
}

var postContents = []string{
	"Recent advancements in reusable rockets have significantly lowered the cost of escaping Earth's gravity, paving the way for upcoming lunar bases and crewed missions to Mars.",
	"Maintaining a consistent sleep schedule and keeping your bedroom cool (around 65°F or 18°C) can dramatically improve the quality of your deep sleep and daily energy levels.",
	"Originally discovered in Ethiopia, coffee cultivation spread to the Arabian Peninsula by the 15th century before becoming the globally consumed beverage we know today.",
	"Often called the eighth wonder of the world, compound interest allows your wealth to grow exponentially as you earn interest on both your initial principal and your accumulated interest.",
	"More than 80% of the world's oceans remain unmapped and unexplored. Extreme pressure and complete darkness make the ocean floor one of the most hostile environments for scientific research.",
	"Modern machine learning algorithms rely on neural networks trained on massive datasets, allowing computers to recognize patterns, translate languages, and generate text with human-like fluency.",
	"Unlike classical computers that use bits representing 0 or 1, quantum computers use qubits which can exist in a state of superposition, enabling them to solve complex equations infinitely faster.",
	"The construction of the Giza Pyramids involved highly organized logistics and skilled laborers who lived in nearby purpose-built villages, debunking old myths about millions of enslaved workers.",
	"Solar panels and wind turbines are seeing exponential global adoption as manufacturing costs fall, shifting global grid dependencies away from traditional fossil fuels.",
	"Plants utilize chlorophyll to capture light energy, converting water and carbon dioxide into glucose and oxygen, providing the primary energy base for almost all life on Earth.",
	"Originating in Japan, this traditional art form transforms a flat sheet of paper into a finished sculpture through specialized folding techniques without utilizing cuts or glue.",
	"By recording transactions across a decentralized network of computers, blockchain technology eliminates the need for trusted third parties like banks to verify financial exchanges.",
	"Magma trapped beneath the Earth's crust contains dissolved gases. As it rises, pressure drops, causing these gases to expand rapidly and trigger violent, explosive eruptions.",
	"Trillions of microbes living inside the human gut form a complex ecosystem that dictates digestion efficiency, metabolic health, and even communicates with the brain via the vagus nerve.",
	"Spanning over 1,400 miles, this massive marine ecosystem is visible from space but faces severe ecological threats from rising ocean temperatures, which cause widespread coral bleaching.",
	"The discovery of linear perspective during the 15th century transformed European painting, allowing artists to create the convincing illusion of three-dimensional depth on flat canvas.",
	"Solid-state batteries are the next major hurdle for electric cars, promising to double current driving ranges while significantly reducing charging times compared to lithium-ion packs.",
	"When massive stars collapse under their own gravity at the end of their life cycles, they create regions of spacetime so dense that even light cannot escape their gravitational pull.",
	"Daily psychological grounding exercises have been shown to physically alter brain structure, shrinking the amygdala to decrease stress responses and lowering baseline cortisol levels.",
	"Vertical hydroponic systems inside empty warehouses allow cities to grow fresh produce year-round, using up to 95% less water than traditional horizontal soil farming.",
}

var postTags = []string{
	"space", "science",
	"health", "wellness", "lifestyle",
	"history", "culture", "beverages",
	"finance", "investing", "education",
	"oceanography", "exploration",
	"technology", "AI", "machine learning",
	"computing", "quantum physics",
	"archaeology", "architecture",
	"energy", "environment", "sustainability",
	"biology", "botany",
}

var postComments = []string{
	"Great insights! I never knew about the advancements in reusable rockets.",
	"This is very informative. I will definitely try to improve my sleep habits.",
	"Fascinating history! I love learning about the origins of coffee.",
	"Thanks for explaining compound interest in such a clear way.",
	"Deep sea exploration is so intriguing! I hope we discover more soon.",
	"AI is evolving so quickly. It's amazing to see how far we've come.",
	"Quantum computing is mind-blowing! I can't wait to see its applications.",
	"Ancient Egypt has always fascinated me. Thanks for sharing these secrets.",
	"Renewable energy is the future. We need to invest more in it.",
	"Photosynthesis is such a fundamental process. Great explanation!",
	"Origami is such a beautiful art form. I love how it transforms paper.",
	"Blockchain technology is revolutionizing finance. Thanks for the insights.",
	"Volcanic eruptions are both terrifying and fascinating. Great article!",
	"The human microbiome is so complex. Thanks for shedding light on it.",
	"The Great Barrier Reef is a natural wonder. We must protect it.",
	"Renaissance art has such a rich history. I love learning about it.",
	"Electric vehicles are the future of transportation. Exciting times ahead!",
	"Black holes are so mysterious. Thanks for explaining them in simple terms.",
	"Mindfulness and meditation have changed my life. Great tips here.",
	"Urban farming is a great solution for sustainable living. Thanks for sharing!",
	"Climate change is a pressing issue. We must act now to protect our planet.",
	"Space exploration is pushing the boundaries of what's possible. Exciting times ahead!",
	"Nutrition is key to a healthy lifestyle. Thanks for the valuable information.",
	"Renewable energy is the future. We need to invest more in it.",
}

func Seed(storage *store.Storage) error {
	ctx := context.Background()
	users := generateUsers(10)
	posts := generatePosts(10)
	comments := generateComments(10)

	for _, user := range users {
		err := storage.Users.Create(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}

	for _, post := range posts {
		err := storage.Posts.Create(ctx, post)
		if err != nil {
			return fmt.Errorf("failed to create post %s: %w", post.Title, err)
		}
	}

	for _, comment := range comments {
		err := storage.Comments.Create(ctx, comment)
		if err != nil {
			return fmt.Errorf("failed to create comment: %w", err)
		}
	}

	fmt.Println("Seeding complete")
	return nil
}

func generateUsers(count int) []*store.User {
	users := make([]*store.User, 0, count)
	for i := 0; i < count; i++ {
		users = append(users, &store.User{
			// ID:       int64(i),
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Password: "password",
		})
	}
	return users
}

func generatePosts(count int) []*store.Post {
	posts := make([]*store.Post, 0, count)
	for i := 0; i < count; i++ {
		posts = append(posts, &store.Post{
			// ID:      int64(i),
			Title:   postTitles[rand.Intn(len(postTitles))],
			Content: postContents[rand.Intn(len(postContents))],
			UserID:  int64(i + 1),
			Tags: []string{
				postTags[rand.Intn(len(postTags))],
				postTags[rand.Intn(len(postTags))],
			},
		})
	}
	return posts
}

func generateComments(count int) []*store.Comment {
	comments := make([]*store.Comment, 0, count)
	for i := 0; i < count; i++ {
		comments = append(comments, &store.Comment{
			// ID:      int64(i),
			Content: postComments[rand.Intn(len(postComments))],
			UserID:  int64(i + 1),
			PostID:  int64(i + 1),
		})
	}
	return comments
}
