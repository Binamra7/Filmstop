import React from "react";
import MovieCardInput from "./MovieCardInput";

const FiveMovies = ({ movies }) => {
	// console.log("five movies: ", movies);
	const handleClick = () => {
		console.log("clicked");
	};
	return (
		<>
			<div
				style={{
					display: "flex",
					justifyContent: "center",
					alignItems: "center",
					marginTop: "50px",
				}}
			>
				<MovieCardInput movie={movies[0]} order="1st" />
			</div>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					justifyContent: "space-evenly",
					alignItems: "center",
					marginTop: "4rem",
				}}
			>
				<MovieCardInput movie={movies[1]} order="2nd" />
				<MovieCardInput movie={movies[2]} order="3rd" />
			</div>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					width: "100%",
					justifyContent: "space-around",
					alignItems: "center",
					marginTop: "4rem",
				}}
			>
				<MovieCardInput movie={movies[3]} order="4th" />
				<MovieCardInput movie={movies[4]} order="5th" />
			</div>
		</>
	);
};

export default FiveMovies;
