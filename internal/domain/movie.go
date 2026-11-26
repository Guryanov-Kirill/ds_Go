package domain

type Movie struct {
	ID       int
	Title    string
	Year     int
	Director string
}

type FailedMovie struct {
	ID      int
	Message string
}

type LoadResult struct {
	Movies []Movie
	Failed []FailedMovie
}
