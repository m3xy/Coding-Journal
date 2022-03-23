import React, { useState } from 'react'
import { Button, FormControl, InputGroup, Toast } from 'react-bootstrap'

function Comment({ID, author, line, b64, replies, show, setShow, replyLine, postReply}) {

    const [text, setText] = useState("");
    const [showReplies, setShowReplies] = useState(false);

    const repliesHTML = (replies !== undefined ? replies.map((reply) => {
        console.log(reply);
        return (<Comment 
            ID={reply.ID} 
            author={reply.author}
            line={reply.lineNumber} 
            b64={reply.base64Value} 
            replies={reply.comments} 
            show={showReplies} 
            setShow={setShowReplies} 
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
                    {atob(b64)}<p/>
                    <InputGroup className="mb-3" size="sm">
                        <FormControl placeholder={"Enter a reply (Line: " + replyLine +  ")"} onChange={(e) => setText(e.target.value)} value={text}/>
                        <Button variant="outline-secondary" onClick={(e) => {setText(""); postReply(e, ID, text, replyLine)}}>Reply</Button>
                    </InputGroup>
                    {replies !== undefined ? 
                        <Button variant='link' size='sm' onClick={() => setShowReplies(!showReplies)}>{showReplies ? <>Hide Replies</> : <>Show Replies</>}</Button>
                    : <></>}      
                </Toast.Body> 
                {repliesHTML}
            </Toast>
        :
            <></>
    )
}

export default Comment