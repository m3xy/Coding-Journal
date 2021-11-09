import React from "react";

class Profile extends React.Component {

	constructor(props) {
        super(props);

        this.state = {
			userId: 0,
            loggedIn: false
        };
    }

	componentDidMount() {
		
	}

	render() {
		//Get user details from ID
		const firstName = "John";
		const lastName = "Doe";
		const userType = "User";
		const email = "JohnDoe@gmail.com"

		//Get user posts
		const posts = []

		//Get user comments
		const comments = []

		//Move to CSS
		const pageCSS = {
			textAlign: "center",
			position: "relative",
			display: "inline-block",
			width: "100%"
		}

		const columnCSS = {
			float: "left",
			width: "50%"
		}

		return (
			// <div style={pageCSS}>
			// 	<header>
			// 		<h2>{firstName + " " + lastName}</h2>
			// 		<div class="dropdown-content">
			// 			<label>({userType})</label>
			// 			<br/>
			// 			<label>Email: {email}</label>
			// 		</div>
			// 	</header>
			// 	<div class="column" style={columnCSS}>
			// 		<div class="dropdown">
			// 			<span><h3>Posts</h3></span>
			// 			<div class="dropdown-content">
			// 				{posts}
			// 			</div>
			// 		</div>
			// 	</div>
			// 	<div class="column" style={columnCSS}>
			// 		<div class="dropdown">
			// 			<span><h3>Comments</h3></span>
			// 			<div class="dropdown-content">
			// 				{comments}
			// 			</div>
			// 		</div>
			// 	</div>
			// </div>
			<div>
				<head>
					<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossOrigin="anonymous"/>
				</head>
				<header>
					<h2>{firstName + " " + lastName}</h2>
			 		<label>({userType})</label>
			 	</header>
				<nav>
					<div className="nav nav-tabs" id="nav-tab" role="tablist">
						<a className="nav-item nav-link active" id="nav-posts-tab" data-toggle="tab" role="tab" aria-controls="nav-home" aria-selected="true">Posts</a>
						<a className="nav-item nav-link" id="nav-comments-tab" data-toggle="tab" role="tab" aria-controls="nav-comments" aria-selected="false">Comments</a>
						<a className="nav-item nav-link" id="nav-contact-tab" data-toggle="tab" role="tab" aria-controls="nav-contact" aria-selected="false">Contact</a>
					</div>
				</nav>
				<div className="tab-content" id="nav-tabContent">
				<div className="tab-pane fade show active" id="nav-home" role="tabpanel" aria-labelledby="nav-posts-tab">{posts}</div>
				<div className="tab-pane fade" id="nav-profile" role="tabpanel" aria-labelledby="nav-comments-tab">{comments}</div>
				<div className="tab-pane fade" id="nav-contact" role="tabpanel" aria-labelledby="nav-contact-tab"><label>Email: {email}</label></div>
				</div>
			</div>
		);
	}
}

export default Profile;