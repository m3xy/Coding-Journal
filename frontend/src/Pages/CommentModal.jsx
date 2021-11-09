import React from 'react'
import {Helmet} from "react-helmet";
import {Modal} from "react-bootstrap";
// import { Modal } from '../_components';


class CommentModal extends React.Component{

    render(){
        return(
			
            <main>
            <h4>Reviewer Comments</h4>
            {/* <Modal show={this.state.show} handleClose={this.hideModal}>
          <p>Modal</p>
        </Modal> */}

            <button type="button" onClick={this.showModal}>
              Add Comment
            </button>
          </main>
			
        )  
    }
}

export default CommentModal;