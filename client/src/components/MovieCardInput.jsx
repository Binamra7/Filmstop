import React, { useState } from "react";
import { Card, Modal } from "@mui/material";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import InputPopup from "./InputPopup";
import { useSelector } from "react-redux";

export default function MediaCardInput({ order, movie, submitted }) {
	const [open, setOpen] = React.useState(false);
	const handleOpen = () => setOpen(true);
	const handleClose = () => setOpen(false);
	const IMG_URL = "https://image.tmdb.org/t/p/w300";
	const source = IMG_URL + movie?.poster_path;
	//https://image.tmdb.org/t/p/w300/nNmJRkg8wWnRmzQDe2FwKbPIsJV.jpg

	return (
		<div>
			<Modal
				open={open}
				onClose={handleClose}
				aria-labelledby="modal-modal-title"
				aria-describedby="modal-modal-description"
				style={{
					display: "flex",
					alignItems: "center",
					justifyContent: "center",
				}}
			>
				<InputPopup order={order} />
			</Modal>
			<Card sx={{ maxWidth: 345 }} onClick={!submitted && handleOpen}>
				<Typography gutterBottom variant="h5" component="div">
					{order}{" "}
				</Typography>
				<CardMedia
					component="img"
					height="400"
					image={
						movie?.title.length > 0
							? source
							: "https://image.shutterstock.com/image-vector/add-icon-new-item-plus-260nw-1315566653.jpg"
					}
					alt="green iguana"
				/>
				<CardContent>
					<Typography gutterBottom variant="h5" component="div">
						{movie?.title}
					</Typography>
				</CardContent>
			</Card>
		</div>
	);
}
