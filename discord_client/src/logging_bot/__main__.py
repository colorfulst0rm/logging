from pocketbase import PocketBase

import json
import discord
import aiohttp
import pocketbase.models
import pocketbase.models.utils

from discord.ext import commands
from discord import app_commands


with open("../config.json", "r") as f:
    config = json.load(f)

with open("../config_secret.json") as f:
    config_secret = json.load(f)

client = commands.Bot(command_prefix=config["prefix"], intents=discord.Intents.all())

pb = PocketBase(config_secret["pb_instance"]["url"])

pb.admins.auth_with_password(config_secret["pb_instance"]["admin_email"], config_secret["pb_instance"]["admin_password"])

def get_guild(guild: discord.Guild):
    try:
        guild_model = pb.collection(b"discord_guilds").get_list(query_params={
            filter: f"snowflake == {guild.id}"
        })
        guild_model = guild_model.items[0]
        
    except:
        guild_model = pb.collection(b"discord_guilds").create({
            "name": guild.name,
            "snowflake": guild.id,
        })
    
    return guild_model

def get_channel(channel: discord.abc.GuildChannel, guild_model: pocketbase.models.utils.base_model.BaseModel):
    try:
        channel_model = pb.collection(b"discord_channels").get_list(query_params={
            filter: f"snowflake == {channel.id}"
        })
        channel_model = channel_model.items[0]
    except:
        channel_model = pb.collection(b"discord_channels").create({
                "name": channel.name,
                "snowflake": channel.id,
                "guild": guild_model.id,
            })
    
    return channel_model
    
def get_user(user: discord.User, guild_model: pocketbase.models.utils.base_model.BaseModel):
    try:
        channel_model = pb.collection(b"discord_users").get_list(query_params={
            filter: f"snowflake == {user.id}"
        })
        channel_model = channel_model.items[0]
    except:
        channel_model = pb.collection(b"discord_channels").create({
            "name": str(user),
            "snowflake": user.id,
            "guild": guild_model.id,
        })

    return channel_model       

@client.event
async def on_message(message: discord.Message):
    guild_model = get_guild(message.guild)
    
    author_model = get_user(message.author, guild_model)

    channel_model  = get_channel(message.channel, guild_model)
    
    embeds = []
    for embed in message.embeds:
        fields = []
        for field in embed.fields:
            fields.append(pb.collection("discord_embed_fields").create({
                "name": field.name,
                "value": field.value,
                "inline": field.inline,
            }))

        embeds.append(pb.collection("discord_embeds").create({
            "title": embed.title,
            "description": embed.description,
            "color": embed.color.value,
            "fields": [field.id for field in fields],
            "url": embed.url,
            "timestamp": embed.timestamp,
            "video": embed.video,
            "author": embed.author,
            "footer_text": embed.footer.text,
            "footer_icon_url": embed.footer.icon_url,
            "provider": embed.provider,
            "type": embed.type,
            "thumbnail": embed.thumbnail.proxy_url,
            "image": embed.image.proxy_url,
        }))

        # with aiohttp.ClientSesspocketbaseion() as session:
        #     with aiohttp.MultipartWriter("form-data") as mp:
                

        # }) as resp:
        #     print(resp.status)
        #     print(await resp.text())


    pb.collection("discord_messages").create({
        "content": message.content,
        "author": author_model.id,
        "channel": channel_model.id,
        "snowflake": message.id,
        "reference": message.reference.message_id if message.reference else None,
        "embeds": [embed.id for embed in embeds],
        "attachments": [pocketbase.models.file_upload.FileUpload((await att.to_file()).fp.read()) for att in message.attachments],
        "guild": guild_model.id,
        })
    
@client.event
async def on_channel_create(channel: discord.abc.GuildChannel):
    if channel.type != discord.ChannelType.text:
        return
    guild_model = pb.collection("discord_guilds").get_one({
        "snowflake": channel.guild.id})

    if not guild_model:
        guild_model = pb.collection("discord_guilds").create({
            "name": channel.guild.name,
            "snowflake": channel.guild.id,
        })

    pb.collection("discord_channels").create({
        "name": channel.name,
        "snowflake": channel.id,
        "guild": guild_model.id,
    })


@client.event
async def on_ready():
    print("finished initialization!")

client.run(config_secret["discord_token"])
