module Gluon  
  class IRC
    def dispatch(msg)
      method = msg['method']
      params = msg['params']
      if method == 'irc.privmsg' && params.has_key?('nick')
        if /^irc$/.match(params['message'])
          response = "usage: irc list|connect <hostname> <nick>|join <channel>"
          Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                                  channel: params['channel'], 
                                                  message: response}})
        end
        cmd_match = /^irc\s+(\w+)(\s+(.*))?/.match(params['message'])
        if cmd_match
          cmd = cmd_match[1]
          cparams = cmd_match[3] ? cmd_match[3].split() : []          
          case cmd
          when 'list'
            do_list(params, cparams)
          when 'connect'
            do_connect(params, cparams)
          when 'join'
            do_join(params, cparams)
          end
        end
      end
    end
 
    def do_list(params, cparams)
      Neur0n.send({id: Neur0n.newId, method: "irc.list"}, "list_resp", params)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on irc list."}})
      return
    end
 
    def list_resp(msg, context)
      vms = msg['result'].map{|v| "##{v['id']}-#{v['state']} #{v['hostname']}/#{v['network']}/#{v['nick']}"}.join(', ')
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: context['irc_session_id'], 
                                                   channel: context['channel'], 
                                                   message: "irc list: #{vms}"}})
      return
    end
 
    def do_join(params, cparams)
      Neur0n.send({id: Neur0n.newId, method: "irc.join", params: {irc_session_id: params['irc_session_id'], 
                                                                  channel: cparams[0] }}, "join_resp", params)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on irc join. #{params} #{cparams}"}})
      return
    end

    def join_resp(msg, context)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: context['irc_session_id'], 
                                                   channel: context['channel'], 
                                                   message: "irc join: #{msg}"}})
      return
    end

    def do_connect(params, cparams)
      icparams = {server: cparams[0], nick: cparams[1] }
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on irc connect. #{params} #{cparams} #{icparams}"}})
      Neur0n.send({id: Neur0n.newId, method: "irc.connect", params: icparams})
      return
    end
   
  end
end
 
module Neur0n
  @@mqueue = {}
  @@client = Gluon::IRC.new
  
  def self.dispatch(msg)
    id = msg['id']
    if id && msg['result']
      q = @@mqueue[id]
      if q
        @@mqueue.delete(q)
        @@client.send(q[:callback], msg, q[:context])
      end
    else
      @@client.dispatch(msg)
    end
  end
 
  def self.send(msg, callback = nil, context = nil)
    if msg[:id]
      @@mqueue[msg[:id]] = {msg: msg, callback: callback, context: context}
    end
    emit(msg)
  end
 
  def self.newId
    (Math.rand*(36**3)).to_i.to_s(36)
  end
end 
