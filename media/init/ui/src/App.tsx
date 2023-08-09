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

type HideableNodes = {
  nodes : Nodes;
  hidden : boolean;
};




// this works when we can overlay over our service.
// should be configurable with a file actually written 
// into this directory
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
function convertNode(p: SNode) : Node {
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
  td.matchesQuery = false;
  //td.context = p.context;
  // XXX hack
  td.derived = ((td.label.indexOf("--")>0) ? true : false);
  td.moderation = a.Moderation ? true : false;
  td.moderationLabel = a.ModerationLabel ? a.ModerationLabel : "";
  td.matchesQuery = false;
  td.children = [];
  return td;
}

function matchTreeState(nodes: Nodes, query: Nodes) {
  Object.keys(nodes).forEach(function(k) {
    nodes[k].matchesQuery = doesMatchQuery(nodes[k],query);
  });
}

// Update the tree state
function convertTreeState(p: SNode, nodes: Nodes):Nodes {
  var n = convertNode(p);
  nodes[n.id] = n;
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertNode(p.children[i])
      nodes[c.id] = c;
      nodes[n.id].children.push(c.id);
    }
  }
  return nodes;
}

function LabeledNode(node: Node) : JSX.Element {
    var thumbnail = node.id+"--thumbnail.png";
    var color="white";
    var textNotMatched = "#a0a0a0";
    if(!node.matchesQuery) {
      color = textNotMatched;
    }
    var note = "";
    if(node.moderation && !node.derived) {
      color = "#ff4040";
      if(!node.matchesQuery) {
        color = "#a03030";
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
    
    var theImg = <></>;
    if(node.id.indexOf("--")<0) {
      theImg = <img 
        src={thumbnail} 
        height="20"
        width="auto" 
        alt="" 
        style={{verticalAlign:'center', border: '0px'}}
        onMouseOver={e => (e.currentTarget.height=200)}
        onMouseOut={e => (e.currentTarget.height=20)}
      />;
    }

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
        color = textNotMatched;
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


function SearchableTreeView() : JSX.Element {
  const [hideData, setHideData] = useState<boolean>(false);
  const [searchData, setSearchData] = useState<Nodes>({});
  const [hideableData, setHideableData] = useState<HideableNodes>({
    hidden: false,
    nodes: {
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
    }
}); 

const detectKeys = async (e : React.KeyboardEvent<HTMLInputElement>) => {
  try {
    if(e.key === "Enter") {
      // Get the keyword and the root to search at
      var searchTerms = e.currentTarget.value;
      var keywords = searchTerms
      var searchAt = "/files/";
      if(searchTerms.startsWith("/files/") ) {
        var tokens = searchTerms.split(" ");
        searchAt = tokens[0];
        keywords = tokens.slice(1).join(" ");
      }
      // Clean the tree before the query
      const response = await fetch(
        endpoint + "/search"+searchAt+"?json=true&hideContent&match="+keywords,
        {"credentials": "same-origin"},
      );
      const p = await response.json() as SNode;
      var existingSearchData = {} as Nodes;
      var newSearchData = convertTreeState(p, existingSearchData);
      setSearchData({...newSearchData});
      matchTreeState(hideableData.nodes,newSearchData);
      setHideableData({...{nodes: (hideableData.nodes), hidden: hideData}});
    }
  } catch (error) {
    console.error('Error fetching data:', error);
  }
  };
  
  const handleClick = async (e: React.MouseEvent<Element,MouseEvent>,node: Node) => {
    try {
      // note: we can delete collapsed nodes to save memory
      if(node.isDir && node.id.endsWith("/")) {
        const response = await fetch(
          endpoint + node.id + "?json=true&listing=true",
          {"credentials": "same-origin"},
        );
        const p = await response.json() as SNode;
        var newTreeData = convertTreeState(p, hideableData.nodes);
        matchTreeState(newTreeData,searchData);
        var newHideableData = {nodes: newTreeData, hidden: hideData};
        setHideableData({...newHideableData});
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };
  
  const detectSelect = async (e: React.SyntheticEvent<Element,Event>) => {
    try {
      console.log("select");
      var newHideData = !hideData;
      setHideData(newHideData);
      var newHideableData = {nodes: hideableData.nodes, hidden: !hideData};
      setHideableData({...newHideableData});
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  var renderTree = function(nodes : Nodes, id:string) : JSX.Element {
    var matches = nodes[id].matchesQuery? true : false;;
    var hidden = !matches && hideData;
    return (
      <TreeItem 
        nodeId={id}
        hidden={hidden} 
        label={LabeledNode(nodes[id])}
        onIconClick={e => handleClick(e,nodes[id])}
        onLabelClick={e => handleClick(e,nodes[id])}
      >
        {Array.isArray(nodes[id].children) ? nodes[id].children.map((v) => renderTree(nodes,v)) : null}
      </TreeItem>
    );
  };

  return (
    <>
    <div style={{padding: 20}}>
    Search like <i>"/files/rob dog OR cat"</i>: &nbsp; <input type="text" name="match" size={40} onKeyUp={e => detectKeys(e)} />
    &nbsp; <input type="checkbox" name="hidemismatch" value="/files/" onClick={e => detectSelect(e)}/> Hide Mismatch
    </div>    
    <TreeView      
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
    >
      {renderTree(hideableData.nodes,"/files/")}
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
        width: 660+400, 
        height: 1000, 
        flexGrow: 0, 
        overflow: 'auto' 
      }}   
    >
      <SearchableTreeView/>
    </div>
  );
}

export default App;
