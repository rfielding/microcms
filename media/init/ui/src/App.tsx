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
  nodes: Node[];
};

const treeData = {
  "id": "/files/",
  "label": "files",
  "nodes": [
    {
      "id": "/files/init/",
      "label": "init",
      "nodes": []
    },
    {
      "id": "/files/rob.fielding@gmail.com/",
      "label": "rob.fielding@gmail.com",
      "nodes": [
        {
          "id": "/files/rob.fielding@gmail.com/permissions.rego",
          "label": "permissions.rego",
          "nodes": []
        },
        {
          "id": "/files/rob.fielding@gmail.com/ktt.jpg",
          "label": "ktt.jpg",
          "nodes": []
        }
      ]
    }
  ]
};

function renderTree(nodes : Node) {
  return (
    <TreeItem key={nodes.id} nodeId={nodes.id} label={nodes.label}>
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
