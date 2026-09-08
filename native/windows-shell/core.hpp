// SPDX-License-Identifier: MIT
// Portable, allocation-bounded menu/parser/handoff logic. No Windows APIs here.
#pragma once
#include <algorithm>
#include <cstdint>
#include <map>
#include <stdexcept>
#include <string>
#include <string_view>
#include <vector>
namespace smolpng  {
    constexpr std::size_t MaxFiles=512, MaxRequest=8*1024*1024;
    struct Action  {
        std::u16string id,label,preset;
        bool interactive=false,inplace=false;
    };
    struct Menu  {
        std::u16string generation,mode,executable,config,home;
        std::vector<Action> actions;
    };
    inline bool identifier(std::u16string_view s)  {
        if(s.empty()||s.size()>64)return false;
        auto alnum=[](char16_t c) {
            return (c>=u'a'&&c<=u'z')||(c>=u'A'&&c<=u'Z')||(c>=u'0'&&c<=u'9');
        };
        if(!alnum(s.front()))return false;
        for(auto c:s)if(!alnum(c)&&c!=u'-'&&c!=u'_')return false;
        return true;
    }
    inline bool generation(std::u16string_view s)  {
        if(s.size()!=32)return false;
        for(auto c:s)if(!((c>=u'0'&&c<=u'9')||(c>=u'a'&&c<=u'f')))return false;
        return true;
    }
    inline bool valid_utf16(std::u16string_view s)  {
        for(std::size_t i=0;i<s.size();++i) {
            auto c=s[i];
            if(!c)return false;
            if(c>=0xd800&&c<=0xdbff) {
                if(i+1==s.size()||s[i+1]<0xdc00||s[i+1]>0xdfff)return false;
                ++i;
            }
            else if(c>=0xdc00&&c<=0xdfff)return false;
        }
        return true;
    }
    inline bool safe_value(std::u16string_view s)  {
        if(!valid_utf16(s))return false;
        for(auto c:s)if(c<u' '||c==127)return false;
        return s.empty()||(s.front()!=u' '&&s.back()!=u' ');
    }
    inline bool supported(std::u16string_view path)  {
        auto at=path.find_last_of(u'.');
        if(at==path.npos)return false;
        std::u16string ext(path.substr(at));
        for(auto &c:ext)if(c>=u'A'&&c<=u'Z')c=char16_t(c+32);
        return ext==u".png"||ext==u".jpg"||ext==u".jpeg";
    }
    inline std::u16string decimal(std::size_t n)  {
        std::u16string s;
        do {
            s.insert(s.begin(),char16_t(u'0'+n%10));
            n/=10;
        }
        while(n);
        return s;
    }
    inline unsigned number(std::u16string_view s,unsigned max)  {
        if(s.empty()||s.size()>5)throw std::runtime_error("invalid number");
        unsigned n=0;
        for(auto c:s) {
            if(c<u'0'||c>u'9')throw std::runtime_error("invalid digit");
            n=n*10+unsigned(c-u'0');
            if(n>max)throw std::runtime_error("number too large");
        }
        return n;
    }
    inline Menu parse_menu(const std::vector<std::uint8_t>& bytes)  {
        if(bytes.size()<2||bytes.size()>65536||bytes.size()%2||bytes[0]!=255||bytes[1]!=254)throw std::runtime_error("invalid menu encoding");
        std::u16string text;
        for(std::size_t i=2;i<bytes.size();i+=2)text.push_back(char16_t(bytes[i]|unsigned(bytes[i+1])<<8));
        if(!valid_utf16(text))throw std::runtime_error("invalid menu text");
        std::map<std::u16string,std::map<std::u16string,std::u16string>> sections;
        std::u16string section;
        for(std::size_t pos=0;pos<text.size();) {
            auto end=text.find(u'\n',pos);
            if(end==text.npos)end=text.size();
            auto line=text.substr(pos,end-pos);
            pos=end+1;
            if(!line.empty()&&line.back()==u'\r')line.pop_back();
            if(line.empty())continue;
            if(line.front()==u'['&&line.back()==u']') {
                section=line.substr(1,line.size()-2);
                if(sections.count(section))throw std::runtime_error("duplicate section");
                sections[section]= {
                };
                continue;
            }
            auto equals=line.find(u'=');
            if(equals==line.npos||section.empty())throw std::runtime_error("invalid INI");
            auto key=line.substr(0,equals);
            if(!sections[section].emplace(key,line.substr(equals+1)).second)throw std::runtime_error("duplicate key");
        }
        auto& m=sections.at(u"menu");
        if(m.size()!=7||m.at(u"schema")!=u"1")throw std::runtime_error("invalid menu schema");
        Menu menu {
            m.at(u"generation"),m.at(u"mode"),m.at(u"executable"),m.at(u"config"),m.at(u"home"), {
            }
        };
        unsigned count=number(m.at(u"count"),16);
        if(!count||sections.size()!=count+1||!generation(menu.generation)||(menu.mode!=u"single"&&menu.mode!=u"submenu")||(menu.mode==u"single"&&count!=1))throw std::runtime_error("invalid menu contract");
        if(!safe_value(menu.executable)||menu.executable.empty()||!safe_value(menu.config)||menu.config.empty()||!safe_value(menu.home))throw std::runtime_error("invalid paths");
        for(unsigned i=0;i<count;i++) {
            auto& a=sections.at(u"action"+decimal(i));
            if(a.size()!=5)throw std::runtime_error("invalid action");
            Action action {
                a.at(u"id"),a.at(u"label"),a.at(u"preset"),number(a.at(u"interactive"),1)==1,number(a.at(u"inplace"),1)==1
            };
            if(!identifier(action.id)||!identifier(action.preset)||!safe_value(action.label)||action.label.empty()||action.label.size()>240||(action.inplace&&action.interactive))throw std::runtime_error("invalid action text");
            for(auto& old:menu.actions)if(old.id==action.id)throw std::runtime_error("duplicate action");
            if((action.inplace&&action.id!=u"replace")||(action.id==u"replace"&&(!action.inplace||action.preset!=u"auto"||action.label!=u"Make smol (replace original)")))throw std::runtime_error("unlabelled destructive action");
            if(menu.mode==u"single"&&(action.id!=u"auto"||action.preset!=u"auto"||action.inplace||action.interactive))throw std::runtime_error("single must preserve original");
            menu.actions.push_back(std::move(action));
        }
        return menu;
    }
    // Quote one Windows argv element; CreateProcessW receives a fixed application path.
    inline std::u16string quote(std::u16string_view s)  {
        std::u16string result=u"\"";
        std::size_t slashes=0;
        for(auto c:s) {
            if(c==u'\\') {
                ++slashes;
                continue;
            }
            result.append(slashes*(c==u'"'?2:1)+(c==u'"'?1:0),u'\\');
            slashes=0;
            result+=c;
        }
        result.append(slashes*2,u'\\');
        result+=u'"';
        return result;
    }
    inline std::vector<std::uint8_t> request(const Menu& menu,const Action& action,const std::vector<std::u16string>& files)  {
        if(!generation(menu.generation)||!identifier(action.id)||files.empty()||files.size()>MaxFiles)throw std::runtime_error("invalid handoff");
        std::string magic="SMOLPNG-SHELL-1\n";
        std::vector<std::uint8_t> b(magic.begin(),magic.end());
        auto u32=[&](std::size_t n) {
            for(int i=0;i<4;++i)b.push_back(std::uint8_t(n>>(i*8)));
        };
        auto string=[&](std::u16string_view s) {
            if(s.empty()||s.size()>32767||!valid_utf16(s))throw std::runtime_error("invalid handoff text");
            u32(s.size());
            for(auto c:s) {
                b.push_back(std::uint8_t(c));
                b.push_back(std::uint8_t(c>>8));
            }
            if(b.size()>MaxRequest)throw std::runtime_error("handoff too large");
        };
        string(menu.generation);
        string(action.id);
        u32(files.size());
        for(auto& f:files) {
            if(!supported(f))throw std::runtime_error("unsupported file");
            string(f);
        }
        return b;
    }
}
