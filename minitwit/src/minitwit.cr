require "kemal"
require "db"
require "sqlite3"

# def connect_db
#   DB.open("minitwit.db")
# end

# before_all do |env|
#   env["db"] = connect_db
#   env["user"] = nil
#   if user_id = env.cookies["user_id"]
#     env["user"] = env["db"].query("select * from user where user_id = ?", [user_id]).first
#   end
# end

# after_all do |env|
#   env["db"].close
# end

get "/" do |context|
  #Shows a users timeline or if no user is logged in it will
  #redirect to the public timeline.  This timeline shows the user"s
  #messages as well as all the messages of followed users.

  puts "We got a visitor from: "
  if true #how to check if valid user
    context.redirect("/public_timeline")
  end
  #offset = context.params.offset || 0
  render("./public/timeline.html" 
    # messages= db.query("
    # select message.*, user.* from message, user
    # where message.flagged = 0 and message.author_id = user.user_id and (
    #     user.user_id = ? or
    #     user.user_id in (select whom_id from follower
    #                             where who_id = ?))
    # order by message.pub_date desc limit ?",
    #[session["user_id"], session["user_id"], PER_PAGE]
    )
end

get "/public_timeline" do |context|
  #Displays the latest messages of all users.
  render("./public/timeline.html"
    # messages=db.query("
    # select message.*, user.* from message, user
    # where message.flagged = 0 and message.author_id = user.user_id
    # order by message.pub_date desc limit ?"#, [PER_PAGE]
    )
end



Kemal.run