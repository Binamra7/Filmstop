export const getMoviesReducer = (state = { movies: [] }, action) => {
	switch (action.type) {
		case "FETCH_MOVIES_REQUEST":
			return {
				// ...state,
				loading: true,
				movies: [],
			};
		case "FETCH_MOVIES_SUCCESS":
			return {
				// ...state,
				loading: false,
				movies: action.payload,
				error: false,
				success: true,
			};
		case "FETCH_MOVIES_FAILURE":
			return {
				// ...state,
				loading: false,
				error: "Failed to fetch movies",
				success: false,
				movies: [],
			};
		default:
			return state;
	}
};

export const searchMovieReducer = (state = { movies: [] }, action) => {
	switch (action.type) {
		case "SEARCH_MOVIE_REQUEST":
			return {
				// ...state,
				loading: true,
			};
		case "SEARCH_MOVIE_SUCCESS":
			return {
				// ...state,
				loading: false,
				movies: action.payload,
				error: false,
				success: true,
			};
		case "SEARCH_MOVIE_FAILURE":
			return {
				// ...state,
				loading: false,
				error: "Failed to fetch movie",
				success: false,
			};
		default:
			return state;
	}
};

export const selectMovieReducer = (state = { movies: [] }, action) => {
	switch (action.type) {
		case "SELECT_MOVIE":
			const newMovie = [...state.movies];
			newMovie[action.payload.order] = action.payload.movie;
			return {
				...state,
				movies: newMovie,
			};
		default:
			return state;
	}
};

export const submitMoviesReducer = (state = { movies: [] }, action) => {
	switch (action.type) {
		case "SUBMIT_MOVIES_REQUEST":
			return {
				// ...state,
				loading: true,
			};
		case "SUBMIT_MOVIES_SUCCESS":
			return {
				// ...state,
				loading: false,
				movies: action.payload,
				error: false,
				success: true,
			};
		case "SUBMIT_MOVIES_FAILURE":
			return {
				// ...state,
				loading: false,
				error: "Failed to submit movies",
				success: false,
			};
		default:
			return state;
	}
};
export const fetchUserMoviesReducer = (state = { movies: [] }, action) => {
	switch (action.type) {
		case "FETCH_USER_MOVIES_REQUEST":
			return {
				loading: true,
				movies: [],
			};
		case "FETCH_USER_MOVIES_SUCCESS":
			return {
				loading: false,
				movies: action.payload,
				error: false,
				success: true,
			};
		case "FETCH_USER_MOVIES_FAILURE":
			return {
				loading: false,
				error: "Failed to fetch movies",
				success: false,
				movies: [],
			};
		default:
			return state;
	}
};
