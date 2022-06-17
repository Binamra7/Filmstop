import { combineReducers } from "redux";
import { createStore, applyMiddleware } from "redux";
import thunk from "redux-thunk";
import {
	getMoviesReducer,
	searchMovieReducer,
	selectMovieReducer,
	submitMoviesReducer,
	fetchUserMoviesReducer,
} from "./reducers/movieReducer";
import { composeWithDevTools } from "redux-devtools-extension";

const finalReducer = combineReducers({
	getMoviesReducer: getMoviesReducer,
	searchMovieReducer: searchMovieReducer,
	selectMovieReducer: selectMovieReducer,
	submitMoviesReducer: submitMoviesReducer,
	fetchUserMoviesReducer: fetchUserMoviesReducer,
});

// const cartItems =
// 	JSON.parse(localStorage.getItem("cartItems")) ||
// 	localStorage.getItem("cartItems") ||
// 	[];

// const currentUser =
// 	JSON.parse(localStorage.getItem("currentUser")) ||
// 	localStorage.getItem("currentUser") ||
// 	null;

// const initialState = {
// 	cartReducer: { cartItems: cartItems },
// 	loginReducer: { currentUser: currentUser },
// };
const initialState = {
	getMoviesReducer: {
		movies: {},
		loading: false,
		error: false,
		success: false,
	},
};
const composeEnhancers = composeWithDevTools({});
const store = createStore(
	finalReducer,
	initialState,
	composeEnhancers(applyMiddleware(thunk))
);

export default store;
