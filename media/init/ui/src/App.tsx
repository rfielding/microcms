import React from 'react';
import logo from './logo.svg';
import './App.css';

import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

interface Node {
  id: string;
  label: string;
  securityLabel: string;
  securityFg: string;
  securityBg: string;
  canRead: boolean;
  canWrite: boolean;
  derived: boolean;
  moderation: boolean;
  moderationLabel: string;
  nodes: Node[];
};

const treeData = {
  "id": "/files/",
  "label": "files",
  "securityLabel": "PUBLIC",
  "securityFg": "white",
  "securityBg": "green",
  "canRead": true,
  "canWrite": true,
  "derived": false,
  "moderation": false,
  "moderationLabel": "",
  "nodes": [
    {
      "id": "/files/init/",
      "label": "init",
      "securityLabel": "PUBLIC",
      "securityFg": "white",
      "securityBg": "green",
      "canRead": true,
      "canWrite": true,
      "derived": false,
      "moderation": false,
      "moderationLabel": "",
      "nodes": []
    },
    {
      "id": "/files/rob.fielding@gmail.com/",
      "label": "rob.fielding@gmail.com",
      "securityLabel": "PUBLIC",
      "securityFg": "white",
      "securityBg": "green",
      "canRead": true,
      "canWrite": true,
      "derived": false,
      "moderation": false,
      "moderationLabel": "",
      "nodes": [
        {
          "id": "/files/rob.fielding@gmail.com/nm.jpg",
          "label": "nm.jpg",
          "securityLabel": "PUBLIC",
          "securityFg": "white",
          "securityBg": "red",
          "canRead": true,
          "canWrite": true,
          "derived": false,
          "moderation": true,
          "moderationLabel": "",
          "nodes": [
            {
              "id": "/files/rob.fielding@gmail.com/nm.jpg--celebs.json",
              "label": "celebs.json",
              "securityLabel": "PUBLIC",
              "securityFg": "white",
              "securityBg": "red",
              "canRead": true,
              "canWrite": true,
              "derived": true,
              "moderation": true,
              "moderationLabel": "female underwear",
              "nodes": []
            }
          ]
        },
        {
          "id": "/files/rob.fielding@gmail.com/ktt.jpg",
          "label": "ktt.jpg",
          "securityLabel": "PUBLIC",
          "securityFg": "white",
          "securityBg": "green",
          "canRead": true,
          "canWrite": true,
          "derived": false,
          "moderation": false,
          "moderationLabel": "",
          "nodes": []
        }
      ]
    }
  ]
};

function labeledNode(node: Node) {
  var opacity = 100;
  if (node.derived) {
    opacity = 25;
  }
  return (
    <>
    <span style={{
      backgroundColor: node.securityBg, 
      color: node.securityFg, 
      opacity: opacity,
    }}>
      {node.securityLabel}&nbsp;{node.canRead ? 'R' : ''}{node.canWrite ? 'W' : ''}{node.moderation ? '!!' : ''}
    </span>
    &nbsp;
    <span>{node.label}</span>
    </>
  );
}

function renderTree(nodes : Node) {
  return (
    <TreeItem key={nodes.id} nodeId={nodes.id} label={labeledNode(nodes)}>
      {Array.isArray(nodes.nodes) ? nodes.nodes.map((node) => renderTree(node)) : null}
    </TreeItem>
  );
}

function MyTreeView() {
  return (
    <TreeView
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
      style={{ alignContent: 'left' ,textAlign: 'left', height: 240, flexGrow: 1, maxWidth: 400, overflowY: 'auto' }}   
    >
      {renderTree(treeData)}
    </TreeView>
  );
}
function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>MicroCMS</h1>
        <MyTreeView/>
      </header>
    </div>
  );
}

export default App;
