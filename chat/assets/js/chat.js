// chat script
// requires jQuery
$(function(){
      "use strict";
      var socket=null,
          msgBox=$('#chatbox #input'),
          messages=$('#messages');
      $('#chatbox').submit(function(e){
          if(!msgBox.val())return false;
          if(!socket){
              alert('Error: no socket connection.');
              return false;
          }
          socket.send(JSON.stringify({"Message":msgBox.val()})); 
          msgBox.val("");
          e.preventDefault();
          e.stopPropagation();
          return false;
      });
      if(!window["WebSocket"]){
          alert('Error: Your browser does not support web sockets.')
      }else{
          socket = new WebSocket("ws://"+host+"/room")
          socket.onopen=function(){
              messages.append(DateFormat(new Date())+" Connection opened.")
          }
          socket.onclose = function(){
              messages.append("Connection has been closed.");
          }
          socket.onmessage=function(e){
			var msg=JSON.parse(e.data);
			messages.append(
				$('<li>').append(
					$('<img>',{src:msg.Avatar,style:"width:3em;"}),
					$('<strong>').text(" "+msg.Name+": "),
					$('<em>').text(DateFormat(new Date(msg.When))),
					$('<span>').text(" "+msg.Message)
				)
			);
          }

      }
})
function ZeroPad(string){
	while(string.toString().length<2){
		string = "0"+string
	}
	return string
}
function DateFormat(date/*Date*/){
	return ZeroPad(date.getHours())+":"+ZeroPad(date.getMinutes())+":"+ZeroPad(date.getSeconds());
}