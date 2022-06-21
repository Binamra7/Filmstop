import axios from "axios";

const herokuAPI = "https://filmstop.herokuapp.com/";

export const getMovies = () => (dispatch) => {
	dispatch({ type: "FETCH_MOVIES_REQUEST" });
	axios
		.get("https://filmstop.herokuapp.com/api/movies")
		.then((res) => {
			dispatch({ type: "FETCH_MOVIES_SUCCESS", payload: res.data });
		})
		.catch((err) => {
			dispatch({ type: "FETCH_MOVIES_FAILURE", payload: err });
		});
};

export const searchMovie = (MOVIE_URL) => (dispatch) => {
	dispatch({ type: "SEARCH_MOVIE_REQUEST" });
	axios
		.get(MOVIE_URL)
		.then((res) => {
			dispatch({ type: "SEARCH_MOVIE_SUCCESS", payload: res.data.results });
		})
		.catch((err) => {
			dispatch({ type: "SEARCH_MOVIE_FAILURE", payload: err });
		});
};

export const selectMovieAction = (movie, order) => {
	// const movies = new Array(5).fill(movie);
	// movies[order] = movie;
	return {
		type: "SELECT_MOVIE",
		payload: { movie, order },
	};
};

export const submitMovies =
	(movies = [], user) =>
	(dispatch) => {
		// console.log("movies: ", movies, typeof movies);

		const formattedMovies = movies.map((movie, index) => {
			return {
				id: movie.id,
				title: movie.title,
				poster_path: movie.poster_path,
				release_date: movie.release_date,
				overview: movie.overview,
			};
		});
		// const formattedMovies = movies;

		dispatch({ type: "SUBMIT_MOVIES_REQUEST" });
		axios
			.post("https://filmstop.herokuapp.com/api/movies/user/add", {
				formattedMovies,
				user,
			})
			.then((res) => {
				dispatch({ type: "SUBMIT_MOVIES_SUCCESS", payload: res.data });
			})
			.catch((err) => {
				dispatch({ type: "SUBMIT_MOVIES_FAILURE", payload: err });
			});
		axios
			.post("https://filmstop.herokuapp.com/api/movies/rank/add", {
				formattedMovies,
				user,
			})
			.then((res) => {
				dispatch({ type: "SUBMIT_MOVIES_SUCCESS", payload: res.data });
			})
			.catch((err) => {
				dispatch({ type: "SUBMIT_MOVIES_FAILURE", payload: err });
			});
	};

export const fetchUserMovies = (user) => (dispatch) => {
	dispatch({ type: "FETCH_USER_MOVIES_REQUEST" });
	axios
		.post("https://filmstop.herokuapp.com/api/movies/user", { user })
		.then((res) => {
			dispatch({ type: "FETCH_USER_MOVIES_SUCCESS", payload: res.data });
		})
		.catch((err) => {
			dispatch({ type: "FETCH_USER_MOVIES_FAILURE", payload: err });
		});
};
