import React from 'react'
import {Helmet} from "react-helmet";
import {Modal, Button, InputGroup, FormControl} from "react-bootstrap";



class CommentModal extends React.Component{

  constructor() {
    super();
    this.state = {
      show: false,
      val: ""
    };
    this.showModal = this.showModal.bind(this);
    this.hideModal = this.hideModal.bind(this);
  }
  showModal = () => {
    this.setState({ show: true });
  }

  hideModal = () => {
    this.setState({ show: false });
  }
  onSubmit = () => {
    this.hideModal
    console.log(this.state.val);
  }

  render() {
    return (
      <>
      <Button variant="primary" onClick={this.showModal}>
        Reviewer Comments
      </Button>

      <Modal show={this.state.show} >
        <Modal.Header closeButton onClick={this.hideModal}>
          <Modal.Title>Comments</Modal.Title>
        </Modal.Header>
        <Modal.Body style={{textAlign: 'center',}}>Enter Reviewer Comments Below
            <InputGroup>
            {/* <InputGroup.Text>Enter Reviewer Comment</InputGroup.Text> */}
            <FormControl as="textarea" aria-label="With textarea" value={this.state.val} onChange={e => this.setState({ val: e.target.value })}
          type="text" />
          </InputGroup>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.hideModal}>
            Close
          </Button>
          <Button variant="primary" onClick={this.onSubmit} >
            Save Changes
          </Button>
        </Modal.Footer>
      </Modal>
    </>
    )
  }

}

export default CommentModal;