module Gluon  
  class VM    
    def dispatch(msg)
      method = msg['method']
      params = msg['params']
      if method == 'irc.privmsg' && params.has_key?('nick')
        if /^vm$/.match(params['message'])
          response = "usage: vm list|add <url>"
          Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                                  channel: params['channel'], 
                                                  message: response}})
        end
        cmd_match = /^vm\s+(\w+)(\s+(.*))?/.match(params['message'])
        if cmd_match
          cmd = cmd_match[1]
          cparams = cmd_match[3] ? cmd_match[3].split() : []          
          case cmd
          when 'list'
            do_list(params, cparams)
          when 'reload'
            do_reload(params, cparams)
          when 'del'
            do_del(params, cparams)
          end
        end
      end
    end

    def do_list(params, cparams)
      Neur0n.send({id: Neur0n.newId, method: "vm.list"}, params)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on vm list."}})
      return
    end

    def resp_list(msg, context)
      vms = msg['result'].values.map{|v| "#{v['name']}"}.join(', ')
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: context['irc_session_id'], 
                                                   channel: context['channel'], 
                                                   message: "vm list: #{vms}"}})
      return
    end

    def do_del(params, cparams)
      Neur0n.send({id: Neur0n.newId, method: "vm.del"}, params)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on vm del. #{params} #{cparams}"}})
      return
    end

    def do_reload(params, cparams)
      Neur0n.send({id: Neur0n.newId, method: "vm.reload", 
                   params: {name: cparams[0]}}, 
                   params)
      Neur0n.send({method: "irc.privmsg", params: {irc_session_id: params['irc_session_id'], 
                                              channel: params['channel'], 
                                              message: "working on vm reload. #{params} #{cparams}"}})
      return
    end

  end
end

module Neur0n
  @@mqueue = {}
  @@client = Gluon::VM.new
  
  def self.dispatch(msg)
    id = msg['id']
    if id && msg['result']
      q = @@mqueue[id]
      if q
        @@mqueue.delete(q)
        @@client.resp_list(msg, q[:context])
      end
    else
      @@client.dispatch(msg)
    end
  end

  def self.send(msg, context = nil)
    if msg[:id]
      @@mqueue[msg[:id]] = {msg: msg, context: context}
    end
    emit(msg)
  end

  def self.newId
    (Math.rand*(36**3)).to_i.to_s(36)
  end
end 