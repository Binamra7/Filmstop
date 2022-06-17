import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";
import { Auth0Provider } from "@auth0/auth0-react";
import store from "./store";
import { Provider } from "react-redux";

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(
	<React.StrictMode>
		<Provider store={store}>
			<Auth0Provider
				domain="dev-cwszvklk.us.auth0.com"
				clientId="Yq9I7hZHC3jZVLq9WaKdO6wM2GAFqx0b"
				redirectUri={window.location.origin}
			>
				<App />
			</Auth0Provider>
		</Provider>
	</React.StrictMode>
);
