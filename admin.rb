puts "I am admin.rb"

module Neur0n
  def self.dispatch(msg)
    puts "I AM DIspatch #{msg.inspect}"
    if msg["code"]
      #Neur0n::add_machine("mac-a")
    end
  end
end
