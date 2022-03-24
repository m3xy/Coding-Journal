/**
 * Comment.jsx
 * Author: 190019931
 * 
 * React component for displaying a comment
 */

import React, { useState } from 'react'
import { Button, Collapse, FormControl, InputGroup, Toast } from 'react-bootstrap'
import JwtService from "../Web/jwt.service";

function Comment({ID, author, line, b64, replies, created, updated, deleted, show, postReply}) {

    const [text, setText] = useState("");
    const [showReplies, setShowReplies] = useState(false);
    const [openReplies, setOpenReplies] = useState(false);

    const repliesHTML = (replies !== undefined ? replies.map((reply) => {
        console.log(reply);
        return (<Comment 
            ID={reply.ID} 
            author={reply.author}
            line={reply.lineNumber} 
            b64={reply.base64Value} 
            replies={reply.comments} 
            created={reply.CreatedAt}
            updated={reply.UpdatedAt}
            deleted={reply.DeletedAt}
            show={showReplies} 
            postReply={postReply}/>)
    })
    : "")

    return (
        show ? 
            <Toast style={{verticalAlign:"top"}} className="d-inline-block m-1">
                <Toast.Header closeButton={false}>
                    {/* <img src="holder.js/20x20?text=%20" className="rounded me-2" alt="" /> */}
                    <strong className="me-auto">{author}</strong>
                    <small>{"Line: " + line}</small>
                </Toast.Header>
                <Toast.Body>
                    {atob(b64)}<p />
                    <small className="text-muted">{deleted ? "(deleted)" : (created == updated ? created : updated + " (edited)")}</small>
                    <br />

                    <Button variant='light' onClick={() => setOpenReplies(!openReplies)}>↩</Button>
                    { JwtService.getUserID() == author ? 
                      <><Button variant='light'>✎</Button>
                        <Button variant='light'>🗑</Button></>
                    : <></>}

                    {replies ? 
                        <Button variant='light' onClick={() => setShowReplies(!showReplies)}>💬</Button>
                    : <></>}

                    <Collapse in={openReplies}>
                        <InputGroup className="mb-3" size="sm">
                            <FormControl placeholder={"Enter a reply"} onChange={(e) => setText(e.target.value)} value={text}/>
                            <Button variant="outline-secondary" onClick={(e) => {setText(""); postReply(e, ID, text)}}>Reply</Button>
                        </InputGroup>
                    </Collapse>
                </Toast.Body> 
                {repliesHTML}
            </Toast>
        :
            <></>
    )
}

export default Comment