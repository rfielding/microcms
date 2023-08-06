import React from 'react';
//import logo from './logo.svg';
import './App.css';

import { useState, useEffect } from 'react';

import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

interface Attributes {
  Label: string;
  LabelFg: string;
  LabelBg: string;
  Read: boolean;
  Write: boolean;
  Derived?: boolean;
  Moderation?: boolean;
  ModerationLabel?: string;
}

interface SNode {
  name: string;
  path: string;
  isDir: boolean;
  size?: number;
  //context?: string;
  part?: number;
  attributes: Attributes;
  children?: SNode[];
};

interface Node {
  id: string;
  label: string;
  isDir: boolean;
  securityLabel: string;
  securityFg: string;
  securityBg: string;
  canRead: boolean;
  canWrite: boolean;
  matchesQuery: boolean;
  part?: number;
  //context?: string;
  derived?: boolean;
  moderation?: boolean;
  moderationLabel?: string;
  children: string[];
};

type Nodes = {
  [id: string]: Node;
};





// this works when we can overlay over our service.
// should be configurable with a file actually written by the user
// inside the unpacked tarball.
var endpoint = "";




function doesMatchQuery(node: Node, query: Nodes) : boolean {
  // cant match empty
  if(Object.keys(query).length === 0) {
    return false;
  }
  // exact match is easy
  if(query[node.id]) {
    return true;
  }
  var answer = false;
  // parent match if our key is a substring of one in query
  Object.keys(query).forEach(function(k) {
    if(k.startsWith(node.id)) {
      answer = true;
    }
  });
  return answer;
}

// Maybe make our json match Material UI's TreeView
function convertNode(p: SNode, query: Nodes) : Node {
  var td = {} as Node;
  td.id = p.path + p.name;
  td.label = p.name;
  if(p.isDir) {
    td.id += "/";
    td.label += "/";
  }
  td.isDir = p.isDir;
  var a = p.attributes;
  td.securityLabel = a.Label;
  td.securityFg = a.LabelFg;
  td.securityBg = a.LabelBg;
  td.canRead = a.Read;
  td.canWrite = a.Write;
  td.part = p.part;
  //td.context = p.context;
  td.derived = a.Derived ? true : false;
  td.moderation = a.Moderation ? true : false;
  td.moderationLabel = a.ModerationLabel ? a.ModerationLabel : "";
  td.matchesQuery = false;
  td.children = [];
  return td;
}

// Update the tree state
function convertTreeState(p: SNode, nodes: Nodes, query: Nodes):Nodes {
  var n = convertNode(p,query);
  nodes[n.id] = n;
  nodes[n.id].matchesQuery = doesMatchQuery(n,query);
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertNode(p.children[i], query)
      nodes[c.id] = c;
      nodes[c.id].matchesQuery = doesMatchQuery(c,query);
      nodes[n.id].children.push(c.id);
    }
  }
  return nodes;
}

function LabeledNode(node: Node) : JSX.Element {
    var thumbnail = node.id+"--thumbnail.png";
    var color="white";
    if(!node.matchesQuery) {
      color = "gray";
    }
    var note = "";
    if(node.moderation && !node.derived) {
      color = "red";
      if(!node.matchesQuery) {
        color = "darkred";
      }
      note = " ( "+node.moderationLabel+" )";
    }
    if(node.derived) {
      thumbnail = "";
      color="gray";
      if(!node.matchesQuery) {
        color = "darkgray";
      }
    }

    var theImg = 
    <img 
      src={thumbnail} 
      height="20"
      width="auto" 
      alt="" 
      style={{verticalAlign:'center'}}
      onMouseOver={e => (e.currentTarget.height=200)}
      onMouseOut={e => (e.currentTarget.height=20)}
      onError={e => (e.currentTarget.onerror = null, e.currentTarget.src = "")}
    />;

    var theText = 
    <a href={node.id} target="_blank" style={{color:color, textDecoration:'none'}}>
      {node.label}&nbsp;
      {note}
      {theImg}
    </a>

    if(node.isDir) {
      thumbnail = "";
      color="white";
      if(!node.matchesQuery) {
        color = "gray";
      }
      theText = 
      <span style={{color:color, textDecoration:'none'}}>
        {node.label}
      </span>;

    }
  
  return (
    <div>
    <span style={{
      backgroundColor: node.securityBg, 
      color: node.securityFg, 
      opacity: 100,
    }}>
      {node.securityLabel}&nbsp;
      {node.canRead ? 'R' : ''}
      {node.canWrite ? 'W' : ''}
      {node.moderation ? '!!' : ''}
    </span>
    &nbsp;
    <span>{theText}</span>
    </div>
  );
};


function FullTreeView() : JSX.Element {
  const [searchData, setSearchData] = useState<Nodes>({
  });

  const detectKeys = async (e:any) => {
    try {
      if(e.key === "Enter") {
        const response = await fetch(
          endpoint + "/search?json=true&match="+e.target.value,
          {"credentials": "same-origin"},
        );
        const p = await response.json() as SNode;
        var newSearchData = convertTreeState(p, searchData, {});
        setSearchData({...newSearchData});
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };


  const [treeData, setTreeData] = useState<Nodes>({
      "/files/": {
        id:"/files/",
        label:"files/",
        isDir:true,
        securityLabel:"PUBLIC",
        securityFg:"white",
        securityBg:"green",
        matchesQuery: false, 
        canRead:true,
        canWrite:false,
        children:[]
      }
  });
  
  const handleToggle = async (node: Node) => {
    try {
      if(node.id.endsWith("/")) {
        const response = await fetch(
          endpoint + node.id + "?json=true&listing=true",
          {"credentials": "same-origin"},
        );
        const p = await response.json() as SNode;
        var newTreeData = convertTreeState(p, treeData,searchData);
        setTreeData({...newTreeData});
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  
  const handleClick = async (node: Node) => {
    if(node.isDir) {
      await handleToggle(node);
    } else {
      console.log("Clicked on "+node.id);
    }
  };
  
  var renderTree = function(nodes : Nodes, id:string) : JSX.Element {
    return (
      <TreeItem 
        nodeId={id} 
        label={LabeledNode(nodes[id])}
        onIconClick={() => handleToggle(nodes[id])}
        onLabelClick={() => handleClick(nodes[id])}
      >
        {Array.isArray(nodes[id].children) ? nodes[id].children.map((v) => renderTree(nodes,v)) : null}
      </TreeItem>
    );
  };

  return (
    <>
    <div style={{padding: 20}}>
    Search: &nbsp; <input type="text" name="match" size={80} onKeyUp={e => detectKeys(e)} />
    </div>    
    <TreeView      
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
    >
      {renderTree(treeData,"/files/")}
    </TreeView>
    </>
  );
}


function App() {
  return (
    <div 
      className="App" 
      style={{ 
        color: 'white', 
        background: 'black', 
        alignContent: 'left', 
        textAlign: 'left', 
        width: 1040, 
        height: 1000, 
        flexGrow: 0, 
        overflow: 'auto' 
      }}   
    >

      <FullTreeView/>
    </div>
  );
}

export default App;
