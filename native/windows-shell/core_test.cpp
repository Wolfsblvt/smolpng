// SPDX-License-Identifier: MIT
#include "core.hpp"
#include <fstream>
#include <iostream>
#include <iterator>
#include <cstdlib>
static void check(bool ok,const char* message){if(!ok){std::cerr<<message<<'\n';std::exit(1);}}
static std::vector<std::uint8_t> ini(std::u16string text){std::vector<std::uint8_t>b{255,254};for(auto c:text){b.push_back(std::uint8_t(c));b.push_back(std::uint8_t(c>>8));}return b;}
int main(int argc,char**argv){
 using namespace smolpng;
 check(quote(uR"(C:\Pictures\a b\)")==u"\"C:\\Pictures\\a b\\\\\"","trailing slash quoting");
 check(quote(u"a\"b")==u"\"a\\\"b\"","quote escaping");
 check(supported(u"C:\\a.JPEG")&&!supported(u"x.png.exe"),"extension selection");
 check(identifier(u"preset-mcp-icon")&&!identifier(u"--evil")&&!identifier(u"a\nb"),"action identifier");
 std::u16string text=u"[menu]\r\nschema=1\r\ngeneration=00112233445566778899aabbccddeeff\r\nmode=single\r\nexecutable=C:\\smolpng\\smolpng.exe\r\nconfig=C:\\User\\config.toml\r\nhome=\r\ncount=1\r\n[action0]\r\nid=auto\r\nlabel=Make smol\r\npreset=auto\r\ninteractive=0\r\ninplace=0\r\n";
 auto menu=parse_menu(ini(text));check(menu.actions.size()==1&&!menu.actions[0].inplace,"single preserves original");
 auto bytes=request(menu,menu.actions[0],{u"C:\\Pictures\\snow \u96ea.png",u"C:\\Pictures\\photo.jpg"});
 check(bytes.size()<MaxRequest,"bounded request");
 if(argc>1){std::ofstream f(argv[1],std::ios::binary);f.write(reinterpret_cast<const char*>(bytes.data()),static_cast<std::streamsize>(bytes.size()));check(bool(f),"write cross-language fixture");}
 bool rejected=false;try{parse_menu(ini(text+u"[action0]\r\nid=other\r\n"));}catch(...){rejected=true;}check(rejected,"duplicate section accepted");
 rejected=false;auto destructive=text;auto at=destructive.find(u"inplace=0");destructive.replace(at,9,u"inplace=1");try{parse_menu(ini(destructive));}catch(...){rejected=true;}check(rejected,"single destructive accepted");
 rejected=false;try{request(menu,menu.actions[0],{});}catch(...){rejected=true;}check(rejected,"empty selection accepted");
 rejected=false;try{request(menu,menu.actions[0],std::vector<std::u16string>(513,u"C:\\a.png"));}catch(...){rejected=true;}check(rejected,"oversized selection accepted");
 rejected=false;auto malformed=ini(text);malformed.push_back(1);try{parse_menu(malformed);}catch(...){rejected=true;}check(rejected,"malformed UTF16 accepted");
 std::cout<<"Native shell contract tests passed\n";
}
