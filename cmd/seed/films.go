package main

import "github.com/Binamra7/filmstop/internal/store"

// A film in the seed catalogue. Ids are the real TMDB ids, so a film picked from
// search later is the same film the seed voted for. appeal is how often voters
// reach for it at all; tags say which tastes reach for it more (see tribes).
type film struct {
	store.Movie
	appeal float64
	tags   []string
}

func f(id int64, title, year, director string, runtime int, genres string, appeal float64, overview string, tags ...string) film {
	return film{
		Movie: store.Movie{ID: id, Title: title, Year: year, Overview: overview,
			Details: store.Details{Director: director, Runtime: runtime, Genres: genres}},
		appeal: appeal, tags: tags,
	}
}

// catalogue spans 1941 to 2024 and leans on a few directors, so the decade, span
// and "top billing" notes all have something to say. Withnail & I is the cult
// film: almost nobody lists it, but the people who do put it first.
var catalogue = []film{
	f(278, "The Shawshank Redemption", "1994", "Frank Darabont", 142, "Drama, Crime", 3, "A banker serving a life sentence finds friendship and a slow, patient way out.", "canon", "crowd"),
	f(238, "The Godfather", "1972", "Francis Ford Coppola", 175, "Drama, Crime", 3, "The ageing head of a crime family hands his empire to the son who never wanted it.", "canon", "classic"),
	f(240, "The Godfather Part II", "1974", "Francis Ford Coppola", 202, "Drama, Crime", 1.8, "Michael tightens his grip on the family while his father's rise plays out decades earlier.", "canon", "classic"),
	f(28, "Apocalypse Now", "1979", "Francis Ford Coppola", 147, "Drama, War", 1.3, "An army captain heads upriver into Cambodia to end a colonel's private war.", "classic", "arthouse"),
	f(155, "The Dark Knight", "2008", "Christopher Nolan", 152, "Action, Crime", 2.5, "Batman, a district attorney and a police lieutenant face a criminal who wants only chaos.", "canon", "crowd"),
	f(27205, "Inception", "2010", "Christopher Nolan", 148, "Action, Science Fiction", 2, "A thief who steals secrets from dreams is hired to plant an idea instead.", "crowd"),
	f(157336, "Interstellar", "2014", "Christopher Nolan", 169, "Adventure, Drama", 2.3, "With Earth failing, a pilot leaves his children to search for a new home among the stars.", "crowd"),
	f(1124, "The Prestige", "2006", "Christopher Nolan", 130, "Drama, Mystery", 1.1, "Two rival stage magicians in Victorian London destroy everything to win.", "crowd", "genre"),
	f(77, "Memento", "2000", "Christopher Nolan", 113, "Mystery, Thriller", 1, "A man who can't form new memories hunts his wife's killer with notes and tattoos.", "genre", "arthouse"),
	f(872585, "Oppenheimer", "2023", "Christopher Nolan", 181, "Drama, History", 1.4, "The physicist who led the Manhattan Project lives with what he built.", "modern", "crowd"),
	f(680, "Pulp Fiction", "1994", "Quentin Tarantino", 154, "Thriller, Crime", 2.4, "Hitmen, a boxer and a gangster's wife cross paths across one long Los Angeles weekend.", "canon", "crowd"),
	f(13, "Forrest Gump", "1994", "Robert Zemeckis", 142, "Comedy, Drama", 1.9, "A kind, simple man drifts through thirty years of American history.", "crowd"),
	f(105, "Back to the Future", "1985", "Robert Zemeckis", 116, "Adventure, Comedy", 1.8, "A teenager sent back to 1955 has to make sure his parents fall in love.", "crowd"),
	f(603, "The Matrix", "1999", "Lana Wachowski & Lilly Wachowski", 136, "Action, Science Fiction", 1.9, "A hacker learns the world he knows is a simulation, and that he may be the one to end it.", "crowd", "genre"),
	f(550, "Fight Club", "1999", "David Fincher", 139, "Drama", 1.9, "An insomniac office worker and a soap salesman start an underground club that gets out of hand.", "crowd"),
	f(769, "GoodFellas", "1990", "Martin Scorsese", 145, "Drama, Crime", 1.7, "Three decades in the life of a mobster, from errand boy to informant.", "canon"),
	f(103, "Taxi Driver", "1976", "Martin Scorsese", 114, "Crime, Drama", 1.2, "A lonely night-shift cabbie in New York decides someone has to clean up the streets.", "classic", "arthouse"),
	f(120, "The Lord of the Rings: The Fellowship of the Ring", "2001", "Peter Jackson", 179, "Adventure, Fantasy", 1.9, "A hobbit inherits a ring of terrible power and sets out to destroy it.", "crowd"),
	f(122, "The Lord of the Rings: The Return of the King", "2003", "Peter Jackson", 201, "Adventure, Fantasy", 2.1, "As the last battle for Middle-earth begins, Frodo and Sam reach Mount Doom.", "crowd"),
	f(11, "Star Wars", "1977", "George Lucas", 121, "Adventure, Action", 1.7, "A farm boy joins a smuggler and a princess against a galactic empire.", "crowd"),
	f(1891, "The Empire Strikes Back", "1980", "Irvin Kershner", 124, "Adventure, Action", 1.9, "The rebels scatter, Luke trains with Yoda, and Vader closes in.", "crowd"),
	f(424, "Schindler's List", "1993", "Steven Spielberg", 195, "Drama, History", 1.9, "A German industrialist saves more than a thousand Jewish workers during the Holocaust.", "canon"),
	f(85, "Raiders of the Lost Ark", "1981", "Steven Spielberg", 115, "Adventure, Action", 1.3, "An archaeologist races the Nazis to the Ark of the Covenant.", "crowd"),
	f(578, "Jaws", "1975", "Steven Spielberg", 124, "Horror, Thriller", 1.2, "A police chief, a scientist and a fisherman hunt a shark off a summer resort town.", "genre", "crowd"),
	f(496243, "Parasite", "2019", "Bong Joon Ho", 133, "Comedy, Thriller", 2.1, "A poor family talks its way, one job at a time, into a rich family's house.", "modern", "canon"),
	f(129, "Spirited Away", "2001", "Hayao Miyazaki", 125, "Animation, Fantasy", 2.3, "A girl trapped in a bathhouse for spirits works to free her parents.", "anim", "canon"),
	f(128, "Princess Mononoke", "1997", "Hayao Miyazaki", 134, "Animation, Fantasy", 1.4, "A cursed prince is caught in a war between a mining town and the forest gods.", "anim"),
	f(8392, "My Neighbor Totoro", "1988", "Hayao Miyazaki", 86, "Animation, Fantasy", 1.2, "Two sisters in the countryside befriend the spirits of the woods.", "anim"),
	f(4935, "Howl's Moving Castle", "2004", "Hayao Miyazaki", 119, "Animation, Fantasy", 1.1, "A young hatter cursed into old age finds work in a wizard's walking castle.", "anim"),
	f(10681, "WALL·E", "2008", "Andrew Stanton", 98, "Animation, Family", 1.3, "The last robot cleaning up an abandoned Earth falls in love.", "anim"),
	f(14160, "Up", "2009", "Pete Docter", 96, "Animation, Comedy", 1.1, "A widower ties balloons to his house and flies to South America, with a stowaway.", "anim"),
	f(862, "Toy Story", "1995", "John Lasseter", 81, "Animation, Adventure", 1.1, "A cowboy doll's place as favourite toy is threatened by a new space ranger.", "anim"),
	f(346648, "Paddington 2", "2017", "Paul King", 104, "Comedy, Family", 0.8, "Paddington is framed for stealing a pop-up book and makes the best of prison.", "anim", "modern"),
	f(289, "Casablanca", "1942", "Michael Curtiz", 102, "Drama, Romance", 1.4, "A nightclub owner in wartime Morocco meets the woman who left him.", "classic"),
	f(15, "Citizen Kane", "1941", "Orson Welles", 119, "Mystery, Drama", 1.2, "Reporters try to decode a newspaper tycoon's dying word.", "classic"),
	f(426, "Vertigo", "1958", "Alfred Hitchcock", 128, "Mystery, Romance", 1.3, "A retired detective with a fear of heights becomes obsessed with the woman he follows.", "classic", "arthouse"),
	f(567, "Rear Window", "1954", "Alfred Hitchcock", 112, "Thriller, Mystery", 1.2, "A photographer laid up with a broken leg suspects a neighbour of murder.", "classic"),
	f(539, "Psycho", "1960", "Alfred Hitchcock", 109, "Horror, Thriller", 1.2, "A secretary on the run stops for the night at a lonely motel.", "classic", "genre"),
	f(389, "12 Angry Men", "1957", "Sidney Lumet", 97, "Drama", 1.7, "One juror holds out against eleven in a murder trial.", "classic", "canon"),
	f(346, "Seven Samurai", "1954", "Akira Kurosawa", 207, "Action, Drama", 1.5, "A village hires seven masterless samurai to defend it from bandits.", "classic", "arthouse"),
	f(18148, "Tokyo Story", "1953", "Yasujirō Ozu", 136, "Drama", 0.9, "An elderly couple visit their grown children in Tokyo, who are too busy for them.", "classic", "arthouse"),
	f(5156, "Bicycle Thieves", "1948", "Vittorio De Sica", 89, "Drama", 0.8, "In postwar Rome, a man and his son search the city for his stolen bicycle.", "classic", "arthouse"),
	f(422, "8½", "1963", "Federico Fellini", 138, "Fantasy, Drama", 0.8, "A blocked film director retreats into memory and fantasy.", "classic", "arthouse"),
	f(935, "Dr. Strangelove or: How I Learned to Stop Worrying and Love the Bomb", "1964", "Stanley Kubrick", 95, "Comedy, War", 1.1, "A rogue general launches a nuclear strike, and the war room can't call it back.", "classic"),
	f(62, "2001: A Space Odyssey", "1968", "Stanley Kubrick", 149, "Science Fiction, Mystery", 1.7, "A black monolith marks humanity's leaps, from the first tools to Jupiter.", "classic", "arthouse", "genre"),
	f(694, "The Shining", "1980", "Stanley Kubrick", 144, "Horror, Thriller", 1.4, "A writer winters with his family as caretaker of an empty mountain hotel.", "genre"),
	f(1398, "Stalker", "1979", "Andrei Tarkovsky", 162, "Science Fiction, Drama", 0.9, "A guide leads two men into the Zone, where a room is said to grant wishes.", "arthouse"),
	f(797, "Persona", "1966", "Ingmar Bergman", 83, "Drama", 0.8, "A nurse cares for an actress who has stopped speaking, and their selves blur.", "arthouse"),
	f(490, "The Seventh Seal", "1957", "Ingmar Bergman", 96, "Fantasy, Drama", 0.8, "A knight back from the Crusades plays chess with Death.", "arthouse", "classic"),
	f(843, "In the Mood for Love", "2000", "Wong Kar-wai", 98, "Drama, Romance", 1.4, "Two neighbours in 1962 Hong Kong discover their spouses are having an affair.", "arthouse"),
	f(1018, "Mulholland Drive", "2001", "David Lynch", 147, "Thriller, Mystery", 1.2, "An aspiring actress and an amnesiac woman piece together a Hollywood mystery.", "arthouse"),
	f(194, "Amélie", "2001", "Jean-Pierre Jeunet", 122, "Comedy, Romance", 1.2, "A shy Parisian waitress quietly rearranges the lives of the people around her.", "arthouse", "crowd"),
	f(598, "City of God", "2002", "Fernando Meirelles", 130, "Drama, Crime", 1.3, "A boy who wants to be a photographer grows up among Rio's gangs.", "canon"),
	f(670, "Oldboy", "2003", "Park Chan-wook", 120, "Drama, Thriller", 1.1, "Released after fifteen years locked in a room, a man has five days to learn why.", "genre", "arthouse"),
	f(38, "Eternal Sunshine of the Spotless Mind", "2004", "Michel Gondry", 108, "Science Fiction, Romance", 1.4, "After a breakup, a man has his ex erased from his memory, and changes his mind halfway.", "arthouse", "modern"),
	f(531428, "Portrait of a Lady on Fire", "2019", "Céline Sciamma", 122, "Drama, Romance", 1.1, "A painter is hired to paint a bride-to-be without her knowing.", "modern", "arthouse"),
	f(545611, "Everything Everywhere All at Once", "2022", "Daniel Kwan & Daniel Scheinert", 140, "Action, Adventure", 1.5, "A laundromat owner fighting a tax audit is pulled across the multiverse.", "modern"),
	f(666277, "Past Lives", "2023", "Celine Song", 106, "Drama, Romance", 1.1, "Childhood sweethearts from Seoul meet again in New York, twenty years on.", "modern"),
	f(965150, "Aftersun", "2022", "Charlotte Wells", 102, "Drama", 1, "A woman revisits a holiday she took with her father when she was eleven.", "modern", "arthouse"),
	f(915935, "Anatomy of a Fall", "2023", "Justine Triet", 152, "Thriller, Mystery", 0.9, "A writer stands trial for her husband's death; their blind son is the key witness.", "modern"),
	f(693134, "Dune: Part Two", "2024", "Denis Villeneuve", 167, "Science Fiction, Adventure", 1.4, "Paul Atreides joins the Fremen and moves towards a war he has foreseen.", "modern", "crowd"),
	f(376867, "Moonlight", "2016", "Barry Jenkins", 111, "Drama", 1, "Three chapters in the life of a young Black man growing up in Miami.", "modern"),
	f(419430, "Get Out", "2017", "Jordan Peele", 104, "Horror, Mystery", 1.2, "A weekend with his girlfriend's parents turns sinister.", "modern", "genre"),
	f(244786, "Whiplash", "2014", "Damien Chazelle", 107, "Drama, Music", 1.7, "A young drummer and a ruthless teacher push each other to the edge.", "modern", "crowd"),
	f(76341, "Mad Max: Fury Road", "2015", "George Miller", 121, "Action, Adventure", 1.4, "A drifter and a warlord's driver flee across the desert with his captive wives.", "genre", "modern"),
	f(348, "Alien", "1979", "Ridley Scott", 117, "Horror, Science Fiction", 1.3, "The crew of a cargo ship answer a distress call and bring something back.", "genre"),
	f(78, "Blade Runner", "1982", "Ridley Scott", 117, "Science Fiction, Drama", 1.4, "A detective in a rain-soaked Los Angeles hunts four escaped replicants.", "genre"),
	f(1091, "The Thing", "1982", "John Carpenter", 109, "Horror, Science Fiction", 0.9, "An Antarctic research team is infiltrated by a creature that imitates its victims.", "genre"),
	f(493922, "Hereditary", "2018", "Ari Aster", 127, "Horror, Mystery", 0.8, "After their grandmother's death, a family uncovers what she left them.", "genre", "modern"),
	f(64690, "Drive", "2011", "Nicolas Winding Refn", 100, "Crime, Drama", 0.9, "A stunt driver who moonlights as a getaway driver takes one job too many.", "genre"),
	f(115, "The Big Lebowski", "1998", "Joel Coen", 117, "Comedy, Crime", 1.1, "A Los Angeles slacker is mistaken for a millionaire and wants his rug back.", "crowd", "genre"),
	f(13446, "Withnail & I", "1987", "Bruce Robinson", 107, "Comedy, Drama", 0.1, "Two unemployed actors escape 1969 London for a disastrous weekend in the country.", "cult"),
}

// tribe is a kind of voter: how much each tag multiplies a film's appeal for
// them, and what share of voters they are.
type tribe struct {
	name  string
	share float64
	boost map[string]float64
}

// tribes give the ballots clusters, so "films that share its ballots" and a
// card's kindred ballot find real neighbours instead of noise.
var tribes = []tribe{
	{"crowd", 0.35, map[string]float64{"crowd": 4, "canon": 3, "anim": 1.5}},
	{"classicist", 0.14, map[string]float64{"classic": 6, "canon": 2.5}},
	{"arthouse", 0.14, map[string]float64{"arthouse": 6, "classic": 1.8, "modern": 1.5}},
	{"animation", 0.09, map[string]float64{"anim": 12}},
	{"now", 0.13, map[string]float64{"modern": 6, "genre": 1.5}},
	{"genre", 0.10, map[string]float64{"genre": 6, "crowd": 1.5}},
	// The cult always opens with Withnail & I (see ballotFor).
	{"cult", 0.05, map[string]float64{"arthouse": 3, "classic": 2}},
}
