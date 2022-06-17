import * as React from "react";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";

export default function MediaCard({ movie, order }) {
	const IMG_URL = "https://image.tmdb.org/t/p/w300";
	const source = IMG_URL + movie.poster_path;
	//https://image.tmdb.org/t/p/w300/nNmJRkg8wWnRmzQDe2FwKbPIsJV.jpg

	const handleClick = () => {
		console.log("clicked");
	};
	return (
		<button onClick={handleClick}>
			<Card sx={{ maxWidth: 345 }}>
				<Typography gutterBottom variant="h5" component="div">
					{order}{" "}
				</Typography>
				<CardMedia
					component="img"
					height="400"
					image={source}
					// image={
					// 	movie.title.length > 0
					// 		? source
					// 		: "https://image.shutterstock.com/image-vector/add-icon-new-item-plus-260nw-1315566653.jpg"
					// }
					alt="green iguana"
				/>
				<CardContent>
					<Typography gutterBottom variant="h5" component="div">
						{movie.title}
					</Typography>
					{/* <Typography variant="body2" color="text.secondary">
					Lizards are a widespread group of squamate reptiles, with over 6,000
					species, ranging across all continents except Antarctica
				</Typography> */}
				</CardContent>
				{/* <CardActions>
				<Button size="small">Share</Button>
				<Button size="small">Learn More</Button>
			</CardActions> */}
			</Card>
		</button>
	);
}
