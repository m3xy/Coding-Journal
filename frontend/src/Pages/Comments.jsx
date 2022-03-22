import React, { useState } from 'react';
import { Button, Form, Modal, Toast } from 'react-bootstrap';
import axiosInstance from '../Web/axiosInstance';
import JwtService from "../Web/jwt.service";

const fileEndpoint = '/file'
const commentEndpoint = '/comment'

function Comments(props) {

	const [show, setShow] = useState(false);
	const [comments, setComments] = useState({
		1:[ {submissionId: null, filePath: null, author: "John Doe", base64Value: "Looks Good!"}, 
			{submissionId: null, filePath: null, author: "Jane Doe", base64Value: "I disagree."},
			{submissionId: null, filePath: null, author: "Jim Doe", base64Value: "I have 500 more citations than both of you, I can assure you, this code is mediocre."}
		  ]
	});
	const [isLoading, setLoading] = useState(false);

	const postComment = (e) => {
		e.preventDefault();
		let comment = document.getElementById("CommentText").value;

		if (comment == null || comment == "") {
			console.log("No comment written");
			return;
		}
		
		let userId = JwtService.getUserID();        //Preparing to get userId from session cookie
		if(userId === null){                        //If user has not logged in, disallow submit
			console.log("Not logged in");
			return;
		}

		// authorId: userId - Should have author?
		// fileId: file.ID
		let data = {
			parentId:null,
			base64Value: btoa(comment)
		}
		console.log(data);
		axiosInstance.post(fileEndpoint + "/" + props.id + commentEndpoint, data)
					.then((response) => {
						console.log(response);
						document.getElementById("CommentText").value = "";
					})
					.catch((error) => {
						console.log(error);
					});
		
		
	}

	const loadMore = () => {
		setLoading(true);
		setTimeout(() => {setLoading(false)}, 1000)
	}

	const commentsHTML = Object.entries(comments).map((line, i) => {
		return line[1].map((comment, j) => {
			return (
				<Toast className="d-inline-block m-1" key={i + ":" + j}>
					<Toast.Header closeButton={false}>
						{/* <img src="holder.js/20x20?text=%20" className="rounded me-2" alt="" /> */}
						<strong className="me-auto">{comment.author}</strong>
						<small>{"Line: " + line[0]}</small>
					</Toast.Header>
					<Toast.Body>{comment.base64Value}</Toast.Body>
				</Toast>
			);
		})
	})

	return (
		<Modal show={props.show} onHide={() => props.setShow(false)} size="lg">
			<Form onSubmit={postComment}>
				<Modal.Header closeButton>
					<Modal.Title>Comments</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					{commentsHTML}
					<div className="d-grid gap-2">
						<Button variant="link" disabled={isLoading} onClick={!isLoading ? () => loadMore() : null}>
							{isLoading ? 'Loading…' : 'Load more'}
						</Button>
					</div>
					<br />
					<Form.Group className="mb-3" controlId="CommentText">
						<Form.Label>Enter a comment below:</Form.Label>
						<Form.Control as="textarea" rows={3} aria-describedby="lineNumber"/>
						<Form.Text id="lineNumber" muted>(Line: {props.line})</Form.Text>
					</Form.Group>              
				</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={() => props.setShow(false)}>
						Close
					</Button>
					<Button variant="primary" type="submit">
						Post comment
					</Button>
				</Modal.Footer>
			</Form>
		</Modal>
	)
}

export default Comments;
