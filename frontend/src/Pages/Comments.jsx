import React, { useState } from 'react';
import { Button, Form, Modal, Toast } from 'react-bootstrap';
import axiosInstance from '../Web/axiosInstance';
import JwtService from "../Web/jwt.service";
import Comment from "./Comment"

const fileEndpoint = '/file'
const commentEndpoint = '/comment'

function Comments({id, comments, setComments, line, show, setShow}) {

	const [text, setText] = useState("");
	const [showComment, setShowComment] = useState(true);
	const [isLoading, setLoading] = useState(false);

	const postComment = (e, parentID, comment) => {
		e.preventDefault();

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
			parentId:parentID,
			base64Value: btoa(comment)
		}
		console.log(data);
		axiosInstance.post(fileEndpoint + "/" + id + commentEndpoint, data)
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

	const commentsHTML = comments.map((comment) => {
		return (<Comment 
					ID={comment.ID} 
					author={comment.author} 
					line={comment.line} 
					b64={comment.base64Value} 
					replies={comment.replies} 
					show={showComment} 
					setShow={setShowComment} 
					replyLine={line} 
					postReply={postComment}/>)
	})

	return (
		<Modal show={show} onHide={() => setShow(false)} size="lg">
			<Form onSubmit={(e) => {setText(""); postComment(e, null, text)}}>
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
						<Form.Control as="textarea" rows={3} aria-describedby="lineNumber" onChange={(e) => {setText(e.target.value)}} value={text}/>
						<Form.Text id="lineNumber" muted>(Line: {line})</Form.Text>
					</Form.Group>              
				</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={() => setShow(false)}>
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
