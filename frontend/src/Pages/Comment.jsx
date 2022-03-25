/**
 * Comment.jsx
 * Author: 190019931
 * 
 * React component for displaying a comment
 */

import React, { useEffect, useState } from 'react'
import { Button, Collapse, FormControl, InputGroup, Toast } from 'react-bootstrap'
import JwtService from "../Web/jwt.service";
import axiosInstance from "../Web/axiosInstance"

const userEndpoint = "/user"

function Comment({comment, show, postReply}) {

    const [reply, setReply] = useState("");
    const [showReplies, setShowReplies] = useState(false);
    const [openReplies, setOpenReplies] = useState(false);
    const [name, setName] = useState("");
    const [edit, setEdit] = useState("");
    const [openEdits, setOpenEdits] = useState(false);


    useEffect(() => {
        getName();
    }, [name])

    const getName = () => {
        axiosInstance.get(userEndpoint + "/" + comment.author)
        .then((response)=>{
            setName(response.data.profile.firstName + " " + response.data.profile.lastName)
        }).catch((error) => {
            console.log(error);
        })
    }

    const postEdit = () => {
        if(text == atob(comment.base64Value)){
            console.log("Comment unchanged")
        }
    }

    const repliesHTML = (comment.comments? comment.comments.map((reply) => {
        return (<Comment 
            comment={reply}
            show={showReplies} 
            postReply={postReply}/>)
    })
    : "")

    return (
        show ? 
            <Toast style={{verticalAlign:"top"}} className="d-inline-block m-1">
                <Toast.Header closeButton={false}>
                    {/* <img src="holder.js/20x20?text=%20" className="rounded me-2" alt="" /> */}
                    <strong className="me-auto">{name}</strong>
                    <small>{"Line: " + comment.lineNumber}</small>
                </Toast.Header>
                <Toast.Body>
                    {openEdits ?
                        <InputGroup className="mb-3" size="sm">
                            <FormControl placeholder={"Edit comment"} onChange={(e) => setEdit(e.target.value)} value={edit}/>
                            <Button variant="outline-secondary" onClick={(e) => {postEdit(); setEdit("")}}>Save</Button>
                        </InputGroup>
                    :
                    atob(comment.base64Value)}<p />
                    <small className="text-muted">{comment.DeletedAt ? "(deleted)" : (comment.CreatedAt == comment.UpdatedAt ? comment.CreatedAt : comment.UpdatedAt + " (edited)")}</small>
                    <br />

                    <Button variant='light' onClick={() => setOpenReplies(!openReplies)}>↩</Button>
                    { JwtService.getUserID() == comment.author ? 
                      <><Button variant='light' onClick={() => {setEdit(atob(comment.base64Value)); setOpenEdits(!openEdits)}}>✎</Button>
                        <Button variant='light'>🗑</Button></>
                    : <></>}

                    {comment.comments ? 
                        <Button variant='light' onClick={() => setShowReplies(!showReplies)}>💬</Button>
                    : <></>}

                    <Collapse in={openReplies}>
                        <InputGroup className="mb-3" size="sm">
                            <FormControl placeholder={"Enter a reply"} onChange={(e) => setReply(e.target.value)} value={reply}/>
                            <Button variant="outline-secondary" onClick={(e) => {postReply(e, comment.ID, reply); setReply("")}}>Reply</Button>
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