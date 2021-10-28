/**
 * Home.jsx
 * 
 * This is the home page of the website. This file holds what will be displayed 
 * when users first access the website
 */

import React from "react";

function Home() {
	return (
		<div className="home">
			<div class="container">
				<div class="row align-items-center my-5">
					<div class="col-lg-7">
						<img
							class="img-fluid rounded mb-4 mb-lg-0"
							src="https://placehold.it/900x400"
							alt=""
						/>
					</div>
					<div class="col-lg-5">
						<h1 class="font-weight-light">Home</h1>
						<p>
							Welcome to CS3099 Group 11 Code Journal!
						</p>
					</div>
				</div>
			</div>
		</div>
	)
}

export default Home;
