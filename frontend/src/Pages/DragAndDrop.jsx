import React, { Component } from 'react'

class DragAndDrop extends Component {
    state = {
        drag: false
    }

    dropRef = React.createRef();

    preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    handleDragStart = (e) => {
        this.preventDefaults(e);
        e.dataTransfer.clearData();
    }

    handleDrag = (e) => {
        this.preventDefaults(e);
    }

    handleDragIn = (e) => {
        this.preventDefaults(e);
        this.dragCounter++;
        if (e.dataTransfer.items && e.dataTransfer.items.length > 0) {
            this.setState({drag: true});
        }
    }

    handleDragOut = (e) => {
        this.preventDefaults(e);
        this.dragCounter--;
        if (this.dragCounter === 0) {
            this.setState({drag: false});
        }
    }

    handleDrop = (e) => {
        this.preventDefaults(e);
        this.setState({drag: false});
        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            this.props.handleDrop(e.dataTransfer.files);
            // e.dataTransfer.clearData();
            this.dragCounter = 0;
        }
    }

    componentDidMount() {
        this.dragCounter = 0;
        let div = this.dropRef.current;
        div.addEventListener('dragstart', this.handleDragStart);
        div.addEventListener('dragenter', this.handleDragIn);
        div.addEventListener('dragleave', this.handleDragOut);
        div.addEventListener('dragover', this.handleDrag);
        div.addEventListener('drop', this.handleDrop);
    }

    componentWillUnmount() {
        let div = this.dropRef.current;
        div.removeEventListener('dragstart', this.handleDragStart);
        div.removeEventListener('dragenter', this.handleDragIn);
        div.removeEventListener('dragleave', this.handleDragOut);
        div.removeEventListener('dragover', this.handleDrag);
        div.removeEventListener('drop', this.handleDrop);
    }

    render() {
        return (
            <div
                style={{display: 'inline-block', position: 'relative'}}
                ref={this.dropRef}
            >
            {this.state.drag &&
                <div 
                    style={{
                    border: 'dashed grey 4px',
                    backgroundColor: 'rgba(255,255,255,.8)',
                    position: 'absolute',
                    top: 0,
                    bottom: 0,
                    left: 0, 
                    right: 0,
                    zIndex: 9999
                    }}
                >
                    <div 
                        style={{
                        position: 'absolute',
                        top: '50%',
                        right: 0,
                        left: 0,
                        textAlign: 'center',
                        color: 'grey',
                        fontSize: 20
                        }}
                    >
                        <div>drop here</div>
                    </div>
                </div>
            }
        {this.props.children}
        </div>
        )
    }
}

export default DragAndDrop