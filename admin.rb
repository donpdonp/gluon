class MasterControlProgram
  def initialize
    @machines = {}
  end

  def dispatch(msg)
    puts "admin.rb dispatch #{msg.inspect}"
    if msg['type'] == 'vm.add'
      if msg['name']
        machine = { id: newId, name: msg['name'], url: msg['url']}
        puts "Adding machine #{machine}"
        idx = Neur0n::machine_add(machine[:id])
        @machines[machine[:id]] = machine
        if idx && machine[:url]
          machine_load(machine)
        end
      end
    end
    if msg['type'] == 'vm.reload'
      machine = @machines[msg['name']]
      machine_load(machine)
    end
    if msg['type'] == 'vm.list'
      #{machines: Neur0n::machine_list}
      puts "list #{@machines.inspect}"
      @machines
    end
  end

  def newId
    alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
    name = ""
    16.times{ name += alphabet[rand(36)]}
    name
  end

  def machine_load(machine)
    url = machine[:url]
    puts "parsing #{url}"
    if gist_id = gistId(url)
      puts "gist id #{gist_id}"
      url = gist_api(gist_id)
    end
    puts "loading #{url}"
    code = Neur0n::http_get(url)
    Neur0n::machine_eval(machine[:id], code)
  end

  def gistId(url)
    gist = url.match(/\/\/gist.github.com\/.*\/(.*)$/)
    return gist[1] if gist
  end

  def gist_api(id)
    gist_api = "https://api.github.com/gists/"+id
    gist = JSON.parse(Neur0n::http_get(gist_api))
    filename = gist['files'].keys.first
    return gist['files'][filename]['raw_url']
  end
end

MCP = MasterControlProgram.new

module Neur0n
  def self.dispatch(msg)
    MCP.dispatch(msg)
  end
end
